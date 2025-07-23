package http

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/config"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/handlers"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/middleware"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/utils"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/image"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/notification"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
	goredis "github.com/redis/go-redis/v9"
)

type ApartmantService struct {
	server           *http.Server
	cfg              *config.Config
	shutdownWG       sync.WaitGroup
	shutdownCtx      context.Context
	cancelFunc       context.CancelFunc
	db               *sqlx.DB
	minioClient      *minio.Client
	redisClient      *goredis.Client
	userHandler      *handlers.UserHandler
	apartmentHandler *handlers.ApartmentHandler
	billHandler      *handlers.BillHandler
}

func NewApartmantService(
	cfg *config.Config,
	db *sqlx.DB,
	minioClient *minio.Client,
	redisClient *goredis.Client,
	userRepo repositories.UserRepository,
	apartmentRepo repositories.ApartmentRepository,
	userApartmentRepo repositories.UserApartmentRepository,
	inviteLinkRepo repositories.InviteLinkFlagRepo,
	notificationService notification.Notification,
	billRepo repositories.BillRepository,
	imageService image.Image,
) *ApartmantService {
	ctx, cancel := context.WithCancel(context.Background())

	userHandler := handlers.NewUserHandler(userRepo)

	apartmentHandler := handlers.NewApartmentHandler(
		apartmentRepo,
		userApartmentRepo,
		inviteLinkRepo,
		notificationService,
		cfg.Server.AppBaseURL,
	)

	billHandler := handlers.NewBillHandler(billRepo, imageService)

	return &ApartmantService{
		cfg:              cfg,
		shutdownCtx:      ctx,
		cancelFunc:       cancel,
		db:               db,
		minioClient:      minioClient,
		redisClient:      redisClient,
		userHandler:      userHandler,
		apartmentHandler: apartmentHandler,
		billHandler:      billHandler,
	}
}

func (s *ApartmantService) Start(serviceName string) error {
	mux := http.NewServeMux()
	s.addCommonRoutes(mux, serviceName)
	s.SetupRoutes(mux)

	s.server = &http.Server{
		Addr:         s.cfg.Server.Port,
		Handler:      ChainMiddleware(mux, middleware.RecoverFromPanic, middleware.LoggingMiddleware, middleware.CorsMiddleware),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.setupSignalHandling()

	s.shutdownWG.Add(1)
	go func() {
		defer s.shutdownWG.Done()

		log.Printf("%s starting on %s", serviceName, s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server failed to start: %v", err)
		}
	}()

	return nil
}

func (s *ApartmantService) methodHandler(methods map[string]http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler, exists := methods[r.Method]
		if !exists {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	}
}

func (s *ApartmantService) setupSignalHandling() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("received signal: %v", sig)

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := s.server.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}

		s.cancelFunc()
	}()
}

func (s *ApartmantService) addCommonRoutes(mux *http.ServeMux, serviceName string) {
	mux.HandleFunc("/health", utils.MethodHandler(map[string]http.HandlerFunc{
		"GET": utils.HealthCheck(serviceName),
	}))
}

func (s *ApartmantService) Stop() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("shutting down server...")
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
		return err
	}
	s.shutdownWG.Wait()

	return nil
}

func (s *ApartmantService) WaitForShutdown() {
	<-s.shutdownCtx.Done()
	s.shutdownWG.Wait()
}

func ChainMiddleware(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for _, mw := range middleware {
		h = mw(h)
	}
	return h
}

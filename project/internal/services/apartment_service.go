package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/config"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/http/middleware"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/repositories"
)

//manager knows which user in which apartment so
//manager invites user x in apartment y

type ApartmentService struct {
	cfg                 *config.Config
	apartmentRepository repositories.ApartmentRepository
	apartmentHandler    *handlers.ApartmentHandler
	server              *http.Server
}

func NewApartmentService(cfg *config.Config) *ApartmentService {
	return &ApartmentService{
		cfg: cfg,
	}
}

func (s *ApartmentService) Start() error {
	//db connection
	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		s.cfg.Postgres.Host, s.cfg.Postgres.Port, s.cfg.Postgres.Username,
		s.cfg.Postgres.Password, s.cfg.Postgres.Database))
	if err != nil {
		return fmt.Errorf("failed to connect to Postgres: %v", err)
	}
	fmt.Println("connected to Postgres")

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	//init repo
	repo, err := repositories.NewApartmentRepository(s.cfg.Postgres.AutoCreate, db)
	if err != nil {
		return fmt.Errorf("failed to create apartment repository: %v", err)
	}
	s.apartmentRepository = repo

	//init handler
	s.apartmentHandler = handlers.NewApartmentHandler(s.apartmentRepository)

	//setting up http server with routes
	mux := http.NewServeMux()
	s.setupRoutes(mux)

	s.server = &http.Server{
		Addr:         "localhost" + s.cfg.Server.Port,
		Handler:      middleware.LoggingMiddleware(middleware.CorsMiddleware(mux)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("starting server on port", s.cfg.Server.Port)
	return s.server.ListenAndServe()
}

func (s *ApartmentService) setupRoutes(mux *http.ServeMux) {
	//grouping manager routes with common prefix and middleware
	managerRoutes := http.NewServeMux()
	managerRoutes.Handle("/apartment/create", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.apartmentHandler.CreateApartment,
	}))
	managerRoutes.Handle("/apartment/get", s.methodHandler(map[string]http.HandlerFunc{
		"GET": s.apartmentHandler.GetApartmentByID,
	}))
	managerRoutes.Handle("/apartment/get-all", s.methodHandler(map[string]http.HandlerFunc{
		"GET": s.apartmentHandler.GetAllApartments,
	}))
	managerRoutes.Handle("/apartment/update", s.methodHandler(map[string]http.HandlerFunc{
		"PUT": s.apartmentHandler.UpdateApartment,
	}))
	managerRoutes.Handle("/apartment/delete", s.methodHandler(map[string]http.HandlerFunc{
		"DELETE": s.apartmentHandler.DeleteApartment,
	}))
	managerRoutes.Handle("/apartment/residents", s.methodHandler(map[string]http.HandlerFunc{
		"GET": s.apartmentHandler.GetResidentsInApartment,
	}))
	managerRoutes.Handle("/apartment/invite", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.apartmentHandler.InviteUserToApartment,
	}))

	//applying manager middleware to all manager routes
	mux.Handle("/api/v1/manager/", middleware.JWTAuthMiddleware(models.Manager)(managerRoutes))

	//grouping resident routes
	residentRoutes := http.NewServeMux()
	residentRoutes.Handle("/apartment/join", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.apartmentHandler.JoinApartment,
	}))
	residentRoutes.Handle("/apartment/leave", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.apartmentHandler.LeaveApartment,
	}))

	//applying resident middleware to all resident routes
	mux.Handle("/api/v1/", middleware.JWTAuthMiddleware(models.Resident)(residentRoutes))

	mux.HandleFunc("/health", s.methodHandler(map[string]http.HandlerFunc{
		"GET": s.healthCheck,
	}))
}

func (s *ApartmentService) methodHandler(methods map[string]http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler, exists := methods[r.Method]
		if !exists {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	}
}

func (s *ApartmentService) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "user-service"}`))
}

func (s *ApartmentService) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("shutting down server...")
	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
		return err
	}

	log.Println("server exited")
	return nil
}

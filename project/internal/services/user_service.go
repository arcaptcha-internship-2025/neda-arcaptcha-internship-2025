package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/config"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/http/handlers"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/http/middleware"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/repositories"
)

type UserService struct {
	cfg            *config.Config
	userRepository repositories.UserRepository
	userHandler    *handlers.UserHandler
	server         *http.Server
}

func NewUserService(cfg *config.Config) *UserService {
	return &UserService{
		cfg: cfg,
	}
}

func (s *UserService) Start() error {
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
	repo, err := repositories.NewUserRepository(s.cfg.Postgres.AutoCreate, db)
	if err != nil {
		return fmt.Errorf("failed to create user repository: %v", err)
	}
	s.userRepository = repo

	//init handler
	s.userHandler = handlers.NewUserHandler(s.userRepository)

	//settingup http server with routes
	mux := http.NewServeMux()
	s.setupRoutes(mux)

	s.server = &http.Server{
		Addr:         "localhost" + s.cfg.Server.Port,
		Handler:      loggingMiddleware(corsMiddleware(mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("server starting on %s", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %v", err)
	}
	return nil
}

func (s *UserService) setupRoutes(mux *http.ServeMux) {
	//first only managers can signup/login
	//manager creates the apartment,and we assume that he knows which user is in which apartments
	//and he sends invitation
	//when the residents accept invitation they are added to user_apartment repo
	//and then residents can sigup and login

	//public routes
	mux.HandleFunc("/api/v1/user/signup", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.userHandler.SignUp,
	}))

	mux.HandleFunc("/api/v1/user/login", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.userHandler.Login,
	}))

	// manager-only routes (requires manager authentication)
	mux.Handle("/api/v1/manager/users", middleware.JWTAuthMiddleware(models.Manager)(
		s.methodHandler(map[string]http.HandlerFunc{
			"GET": s.userHandler.GetAllUsers,
		}),
	))

	mux.Handle("/api/v1/manager/users/", middleware.JWTAuthMiddleware(models.Manager)(
		s.methodHandler(map[string]http.HandlerFunc{
			"GET":    s.userHandler.GetUser,
			"DELETE": s.userHandler.DeleteUser,
		}),
	))

	// user routes (requires authentication for both manager and resident)
	mux.Handle("/api/v1/user/profile", middleware.JWTAuthMiddleware(models.Resident)(
		s.methodHandler(map[string]http.HandlerFunc{
			"GET": s.userHandler.GetProfile,
			"PUT": s.userHandler.UpdateProfile,
		}),
	))

	mux.Handle("/api/v1/user/profile/picture", middleware.JWTAuthMiddleware(models.Resident)(
		s.methodHandler(map[string]http.HandlerFunc{
			"POST": s.userHandler.UploadProfilePicture,
		}),
	))

	// health check (public)
	mux.HandleFunc("/health", s.methodHandler(map[string]http.HandlerFunc{
		"GET": s.healthCheck,
	}))
}

// wraps handlers to support different http methods
func (s *UserService) methodHandler(methods map[string]http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler, exists := methods[r.Method]
		if !exists {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	}
}

// makes a simple health check endpoint
func (s *UserService) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "user-service"}`))
}

// shuts down the server
func (s *UserService) Stop() error {
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

// logs all incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s %s - Started", r.RemoteAddr, r.Method, r.URL.Path)

		//custom ResponseWriter to capture status code
		ww := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		duration := time.Since(start)
		log.Printf("%s %s %s - Completed in %v with status %d",
			r.RemoteAddr, r.Method, r.URL.Path, duration, ww.statusCode)
	})
}

// handles CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// captures the status code for logging
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWrapper) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

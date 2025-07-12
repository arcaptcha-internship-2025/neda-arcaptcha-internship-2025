package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/config"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/app"
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
	db, err := app.ConnectToDatabase(s.cfg.Postgres)
	if err != nil {
		return fmt.Errorf("failed to connect to Postgres: %v", err)
	}

	repo, err := repositories.NewUserRepository(s.cfg.Postgres.AutoCreate, db)
	if err != nil {
		return fmt.Errorf("failed to create user repository: %v", err)
	}
	s.userRepository = repo

	s.userHandler = handlers.NewUserHandler(s.userRepository)

	mux := http.NewServeMux()
	s.setupRoutes(mux)

	s.server = &http.Server{
		Addr:         "localhost" + s.cfg.Server.Port,
		Handler:      middleware.LoggingMiddleware(middleware.CorsMiddleware(mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("User service starting on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *UserService) setupRoutes(mux *http.ServeMux) {
	api := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	//public routes
	api.HandleFunc("/user/signup", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.userHandler.SignUp,
	}))

	api.HandleFunc("/user/login", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.userHandler.Login,
	}))

	//grouping manager routes
	managerRoutes := http.NewServeMux()
	api.Handle("/manager/", http.StripPrefix("/manager", middleware.JWTAuthMiddleware(models.Manager)(managerRoutes)))

	managerRoutes.HandleFunc("/user/get-all", s.methodHandler(map[string]http.HandlerFunc{
		"GET": s.userHandler.GetAllUsers,
	}))

	managerRoutes.HandleFunc("/user/get", s.methodHandler(map[string]http.HandlerFunc{
		"GET": s.userHandler.GetUser,
	}))

	managerRoutes.HandleFunc("/user/delete", s.methodHandler(map[string]http.HandlerFunc{
		"DELETE": s.userHandler.DeleteUser,
	}))

	//grouping resident routes
	residentRoutes := http.NewServeMux()
	api.Handle("/resident/", http.StripPrefix("/resident", middleware.JWTAuthMiddleware(models.Resident)(residentRoutes)))

	residentRoutes.HandleFunc("/profile", s.methodHandler(map[string]http.HandlerFunc{
		"GET": s.userHandler.GetProfile,
		"PUT": s.userHandler.UpdateProfile,
	}))

	residentRoutes.HandleFunc("/profile/picture", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.userHandler.UploadProfilePicture,
	}))

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

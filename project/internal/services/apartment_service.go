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
	db, err := app.ConnectToDatabase(s.cfg.Postgres)
	if err != nil {
		return fmt.Errorf("failed to connect to Postgres: %v", err)
	}

	// init repo
	repo, err := repositories.NewApartmentRepository(s.cfg.Postgres.AutoCreate, db)
	if err != nil {
		return fmt.Errorf("failed to create apartment repository: %v", err)
	}
	s.apartmentRepository = repo

	// init handler
	s.apartmentHandler = handlers.NewApartmentHandler(s.apartmentRepository)

	// setting up http server with routes
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
	api := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	//grouping manager routes
	managerRoutes := http.NewServeMux()
	api.Handle("/manager/", http.StripPrefix("/manager", middleware.JWTAuthMiddleware(models.Manager)(managerRoutes)))
	managerRoutes.HandleFunc("/apartment/create", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.apartmentHandler.CreateApartment,
	}))
	managerRoutes.HandleFunc("/apartment/get", s.methodHandler(map[string]http.HandlerFunc{
		"GET": s.apartmentHandler.GetApartmentByID,
	}))
	managerRoutes.HandleFunc("/apartment/get-all", s.methodHandler(map[string]http.HandlerFunc{
		"GET": s.apartmentHandler.GetAllApartments,
	}))
	managerRoutes.HandleFunc("/apartment/update", s.methodHandler(map[string]http.HandlerFunc{
		"PUT": s.apartmentHandler.UpdateApartment,
	}))
	managerRoutes.HandleFunc("/apartment/delete", s.methodHandler(map[string]http.HandlerFunc{
		"DELETE": s.apartmentHandler.DeleteApartment,
	}))
	managerRoutes.HandleFunc("/apartment/residents", s.methodHandler(map[string]http.HandlerFunc{
		"GET": s.apartmentHandler.GetResidentsInApartment,
	}))
	managerRoutes.HandleFunc("/apartment/invite", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.apartmentHandler.InviteUserToApartment,
	}))

	//grouping resident routes
	residentRoutes := http.NewServeMux()
	api.Handle("/resident/", http.StripPrefix("/resident", middleware.JWTAuthMiddleware(models.Resident)(residentRoutes)))
	residentRoutes.HandleFunc("/apartment/join", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.apartmentHandler.JoinApartment,
	}))
	residentRoutes.HandleFunc("/apartment/leave", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.apartmentHandler.LeaveApartment,
	}))

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

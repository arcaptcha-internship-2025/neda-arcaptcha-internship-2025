package http

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/config"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/http/middleware"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/http/utils"
)

type RouteSetup interface {
	SetupRoutes(mux *http.ServeMux)
}

type HTTPServer struct {
	server *http.Server
	cfg    *config.Config
}

func NewHTTPServer(cfg *config.Config) *HTTPServer {
	return &HTTPServer{
		cfg: cfg,
	}
}

func (s *HTTPServer) Start(routeSetup RouteSetup, serviceName string) error {
	mux := http.NewServeMux()

	routeSetup.SetupRoutes(mux)

	s.addCommonRoutes(mux, serviceName)

	s.server = &http.Server{
		Addr:         "localhost" + s.cfg.Server.Port,
		Handler:      middleware.LoggingMiddleware(middleware.CorsMiddleware(mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("%s starting on %s", serviceName, s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *HTTPServer) addCommonRoutes(mux *http.ServeMux, serviceName string) {
	mux.HandleFunc("/health", utils.MethodHandler(map[string]http.HandlerFunc{
		"GET": utils.HealthCheck(serviceName),
	}))
}

func (s *HTTPServer) Stop() error {
	if s.server == nil {
		return nil
	}

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

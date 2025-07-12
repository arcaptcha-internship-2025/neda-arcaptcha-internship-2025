package http

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/http/middleware"
)

type Server struct {
	*http.Server
}

func NewServer(addr string, setupRoutes func(*http.ServeMux)) *Server {
	mux := http.NewServeMux()
	setupRoutes(mux)

	return &Server{
		&http.Server{
			Addr:         addr,
			Handler:      middleware.LoggingMiddleware(middleware.CorsMiddleware(mux)),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func (s *Server) Start(serviceName string) error {
	log.Printf("%s starting on %s", serviceName, s.Addr)
	return s.ListenAndServe()
}

func (s *Server) Stop(serviceName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("shutting down %s server...", serviceName)
	if err := s.Shutdown(ctx); err != nil {
		log.Printf("%s server forced to shutdown: %v", serviceName, err)
		return err
	}

	log.Printf("%s server exited", serviceName)
	return nil
}

func MethodHandler(methods map[string]http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler, exists := methods[r.Method]
		if !exists {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	}
}

func HealthCheck(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "` + serviceName + `"}`))
	}
}

package utils

import (
	"encoding/json"
	"net/http"
)

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

		response := map[string]string{
			"status":  "healthy",
			"service": serviceName,
		}

		json.NewEncoder(w).Encode(response)
	}
}

// creates a new ServeMux with API prefix handling
func APIPrefix(mux *http.ServeMux) *http.ServeMux {
	api := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))
	return api
}

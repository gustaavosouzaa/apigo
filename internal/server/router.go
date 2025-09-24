package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"apigo/internal/geocode"
)

// RegisterRoutes configures the HTTP handlers for the service.
func RegisterRoutes(mux *http.ServeMux, service *geocode.Service) {
	mux.HandleFunc("/geocode", geocodeHandler(service))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
}

func geocodeHandler(service *geocode.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		address := strings.TrimSpace(r.URL.Query().Get("address"))
		if address == "" {
			respondError(w, http.StatusBadRequest, "address query parameter is required")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		result, err := service.Geocode(ctx, address)
		if err != nil {
			switch {
			case errors.Is(err, geocode.ErrNoResults):
				respondError(w, http.StatusNotFound, err.Error())
			case errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled):
				respondError(w, http.StatusGatewayTimeout, "geocoding request timed out")
			default:
				respondError(w, http.StatusBadGateway, err.Error())
			}
			return
		}

		respondJSON(w, http.StatusOK, result)
	}
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

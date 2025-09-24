package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"apigo/internal/config"
	"apigo/internal/geocode"
	"apigo/internal/server"
)

func main() {
	if err := config.LoadFromEnvFile(".env"); err != nil {
		log.Fatalf("failed to load .env file: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	service := geocode.NewService(cfg.GoogleAPIKey, 30*time.Minute)

	mux := http.NewServeMux()
	server.RegisterRoutes(mux, service)

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("starting server on port %s", cfg.ServerPort)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server failed: %v", err)
	}
}

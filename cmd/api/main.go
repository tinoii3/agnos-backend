package main

import (
	"log"

	"agnos-backend/internal/config"
	"agnos-backend/internal/db"
	"agnos-backend/internal/http"
	"agnos-backend/internal/store"
)

func main() {
	cfg := config.Load()

	dbPool, err := db.NewPostgresPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to initialize postgres: %v", err)
	}
	defer dbPool.Close()

	st := store.NewPostgresStore(dbPool)
	server := http.NewServer(st, cfg.JWTSecret)
	router := server.Router()

	log.Printf("server starting on %s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

package main

import (
	"log"

	"agnos-backend/internal/config"
	"agnos-backend/internal/db"
	"agnos-backend/internal/http"
)

func main() {
	cfg := config.Load()

	dbPool, err := db.NewPostgresPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to initialize postgres: %v", err)
	}
	defer dbPool.Close()

	router := http.NewRouter(dbPool)

	log.Printf("server starting on %s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

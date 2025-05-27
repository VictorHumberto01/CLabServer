package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"clab-server/internal/api/middleware"
	"clab-server/internal/api/routes"
	"clab-server/internal/config"
	"clab-server/internal/database/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := db.AutoMigrate(
		&models.User{},
		&models.Room{},
		&models.Task{},
		&models.TestCase{},
		&models.Submission{},
	); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)

	// Setup routes
	router := routes.SetupRoutes(db, authMiddleware, cfg)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

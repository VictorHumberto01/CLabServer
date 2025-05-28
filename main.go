package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/vitub/CLabServer/pkg/api"
	"github.com/vitub/CLabServer/pkg/banner"
	"github.com/vitub/CLabServer/pkg/security"
)

func main() {
	// Display the cLab banner
	banner.PrintBanner()

	// Check if GCC is available
	if !security.IsCommandAvailable("gcc") {
		log.Fatal("GCC compiler not found. Please install GCC to use this server.")
	}

	// Check if firejail is available
	if !security.IsCommandAvailable("firejail") {
		if !security.PromptForUnsecureMode() {
			log.Fatal("Firejail not available and unsecure mode rejected")
		}
		log.Println("⚠️  Running in UNSECURE mode - code execution is not sandboxed!")
	} else {
		log.Println("✅ Firejail detected - code execution will be sandboxed")
	}

	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	r := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// Setup routes
	api.SetupRoutes(r)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("🚀 Server starting on port %s", port)
	log.Println("📡 Endpoints available:")
	log.Println("   GET  /health  - Health check")
	log.Println("   POST /compile - Compile and run C code")

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/vitub/CLabServer/internal/api/routes"
	"github.com/vitub/CLabServer/internal/banner"
	"github.com/vitub/CLabServer/internal/initializers"
	"github.com/vitub/CLabServer/internal/security"
)

func main() {
	// Initialize Environment Variables
	initializers.LoadEnvVariables()

	// Initialize Database
	if err := initializers.ConnectToDB(); err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	banner.PrintBanner()

	if !security.IsCommandAvailable("gcc") {
		log.Fatal("GCC compiler not found. Please install GCC to use this server.")
	}

	if !security.IsCommandAvailable("firejail") {
		if !security.PromptForUnsecureMode() {
			log.Fatal("Firejail not available and unsecure mode rejected")
		}
		log.Println("⚠️  Running in UNSECURE mode - code execution is not sandboxed!")
	} else {
		log.Println("✅ Firejail detected - code execution will be sandboxed")
	}

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	routes.SetupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)
	log.Println("📡 Endpoints available:")
	log.Println("   GET  /health  - Health check")
	log.Println("   POST /compile - Compile and run C code")
	log.Println("   POST /signup - Sign up")
	log.Println("   POST /login - Login")
	log.Println("   POST /login/cookie - Login with cookie")
	log.Println("   GET /validate - Validate token")

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

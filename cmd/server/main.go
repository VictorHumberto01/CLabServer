package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/vitub/CLabServer/internal/api/handlers"
	"github.com/vitub/CLabServer/internal/api/middleware"
	"github.com/vitub/CLabServer/internal/api/routes"
	"github.com/vitub/CLabServer/internal/banner"
	"github.com/vitub/CLabServer/internal/initializers"
	"github.com/vitub/CLabServer/internal/security"
	"github.com/vitub/CLabServer/internal/ws"
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
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	hub := ws.NewHub()
	go hub.Run()

	r.GET("/ws", middleware.OptionalAuth, func(c *gin.Context) {
		ws.ServeWs(hub, c)
	})

	r.POST("/admin/create-teacher", handlers.CreateTeacher)

	routes.SetupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)
	log.Println("📡 Endpoints available:")
	endpoints := []struct {
		method string
		path   string
		desc   string
	}{
		{"GET", "/health", "Health check"},
		{"GET", "/ws", "WebSocket Endpoint"},
		{"POST", "/compile", "Compile and run C code"},

		{"POST", "/login", "Login"},
		{"POST", "/login/cookie", "Login with cookie"},
		{"GET", "/validate", "Validate token"},
		{"POST", "/classrooms", "Create classroom"},
		{"GET", "/classrooms", "List classrooms"},
		{"POST", "/classrooms/:id/students", "Add student to classroom"},
		{"GET", "/history", "List history"},
		{"POST", "/admin/create-teacher", "Create Teacher (Admin)"},
	}

	for _, e := range endpoints {
		log.Printf("   %s %s - %s", e.method, e.path, e.desc)
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

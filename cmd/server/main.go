package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/vitub/CLabServer/internal/api/routes"
	"github.com/vitub/CLabServer/internal/banner"
	"github.com/vitub/CLabServer/internal/initializers"
	"github.com/vitub/CLabServer/internal/ws"
)

func checkDocker() error {
	cmd := exec.Command("docker", "info")
	return cmd.Run()
}

func isRunningInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}

func main() {
	if err := checkDocker(); err != nil {
		log.Fatalf("🚨 ERROR: Docker is not running or not installed!\n"+
			"This server requires Docker to run the C code compilation sandbox.\n"+
			"Please install and start Docker, or use 'docker-compose up -d' to run the backend properly.\n\n"+
			"Details: %v", err)
	}

	initializers.LoadEnvVariables()

	if err := initializers.ConnectToDB(); err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	banner.PrintBanner()

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

	routes.SetupRoutes(r, hub)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	banner.PrintRoutes(r)
	banner.PrintStartup(port)

	if !isRunningInContainer() {
		log.Println("⚠️ SECURITY WARNING:\n" +
			"Running the server outside a Docker container (e.g., using 'go run') makes the API\n" +
			"vulnerable to host-level attacks. While the C code sandbox remains perfectly isolated,\n" +
			"any vulnerability in the Go HTTP API could compromise your host machine's file system\n" +
			"and resources. Using docker-compose is highly recommended for full isolation.")
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/vitub/CLabServer/internal/api/routes"
	"github.com/vitub/CLabServer/internal/banner"
	"github.com/vitub/CLabServer/internal/initializers"
	"github.com/vitub/CLabServer/internal/ws"
)

func main() {
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

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

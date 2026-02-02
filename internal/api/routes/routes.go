package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vitub/CLabServer/internal/api/handlers"
)

// SetupRoutes configures all the routes for the API
func SetupRoutes(r *gin.Engine) {
	// Add middleware for logging
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Compilation endpoint
	r.POST("/compile", handlers.HandleCompile)

	// CORS preflight
	r.OPTIONS("/compile", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
}

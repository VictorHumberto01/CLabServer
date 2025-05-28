package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vitub/CLabServer/pkg/compiler"
	"github.com/vitub/CLabServer/pkg/models"
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
	r.POST("/compile", handleCompile)

	// CORS preflight
	r.OPTIONS("/compile", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
}

// handleCompile handles the compilation request
func handleCompile(c *gin.Context) {
	var req models.CompileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid request format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format. Expected JSON with 'code' field.",
		})
		return
	}

	// Basic validation
	if strings.TrimSpace(req.Code) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Code cannot be empty",
		})
		return
	}

	// Validate timeout
	if req.TimeoutSecs > 30 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Timeout cannot exceed 30 seconds",
		})
		return
	}

	// Log the compilation request
	inputInfo := ""
	if len(req.InputLines) > 0 {
		inputInfo = fmt.Sprintf(", input lines: %d", len(req.InputLines))
	} else if req.Input != "" {
		inputInfo = fmt.Sprintf(", input length: %d", len(req.Input))
	}
	log.Printf("Received compilation request, code length: %d%s", len(req.Code), inputInfo)

	// Compile and run
	response := compiler.CompileAndRun(req)

	// Log the response
	if response.Error != "" {
		log.Printf("Compilation/execution failed: %s", response.Error)
	} else {
		log.Printf("Compilation successful, output length: %d", len(response.Output))
	}

	c.JSON(http.StatusOK, response)
}

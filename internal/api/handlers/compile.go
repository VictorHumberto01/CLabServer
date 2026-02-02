package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vitub/CLabServer/internal/compiler"
	"github.com/vitub/CLabServer/internal/models"
)

// HandleCompile handles the compilation request
func HandleCompile(c *gin.Context) {
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

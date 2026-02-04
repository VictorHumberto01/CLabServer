package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vitub/CLabServer/internal/compiler"
	"github.com/vitub/CLabServer/internal/initializers"
	"github.com/vitub/CLabServer/internal/models"
)

func HandleCompile(c *gin.Context) {

	user, _ := c.Get("user")

	var req models.CompileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid request format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format. Expected JSON with 'code' field.",
		})
		return
	}

	if strings.TrimSpace(req.Code) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Code cannot be empty",
		})
		return
	}

	if req.TimeoutSecs > 30 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Timeout cannot exceed 30 seconds",
		})
		return
	}

	inputInfo := ""
	if len(req.InputLines) > 0 {
		inputInfo = fmt.Sprintf(", input lines: %d", len(req.InputLines))
	} else if req.Input != "" {
		inputInfo = fmt.Sprintf(", input length: %d", len(req.Input))
	}
	log.Printf("Received compilation request, code length: %d%s", len(req.Code), inputInfo)

	response := compiler.CompileAndRun(req)

	// Log the response
	if response.Error != "" {
		log.Printf("Compilation/execution failed: %s", response.Error)
	} else {
		log.Printf("Compilation successful, output length: %d", len(response.Output))
	}

	// Save history if user is authenticated
	if user != nil {
		if u, ok := user.(models.User); ok {
			history := models.History{
				UserID: u.ID,
				Code:   req.Code,
				Input:  req.Input,
				Output: response.Output,
				Error:  response.Error,
			}
			initializers.DB.Create(&history)
			log.Printf("User was logged in and history was saved")
		} else {
			log.Printf("User was not logged in")
		}
	}

	c.JSON(http.StatusOK, response)
}

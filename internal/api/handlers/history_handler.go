package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vitub/CLabServer/internal/initializers"
	"github.com/vitub/CLabServer/internal/models"
)

func ListHistory(c *gin.Context) {
	user, _ := c.Get("user")
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	u, ok := user.(models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user session"})
		return
	}
	var history []models.History
	initializers.DB.Where("user_id = ?", u.ID).Find(&history)

	c.JSON(http.StatusOK, history)

}

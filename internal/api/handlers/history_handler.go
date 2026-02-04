package handlers

import (
	"fmt"
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

	pageVal := c.DefaultQuery("page", "1")
	limitVal := c.DefaultQuery("limit", "10")
	search := c.Query("search")

	page := 1
	limit := 10

	if p, err := parseUint(pageVal); err == nil && p > 0 {
		page = p
	}
	if l, err := parseUint(limitVal); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	offset := (page - 1) * limit

	var history []models.History
	var total int64

	query := initializers.DB.Model(&models.History{}).Where("user_id = ?", u.ID)

	if search != "" {
		query = query.Where("code LIKE ?", "%"+search+"%")
	}

	query.Count(&total)

	result := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&history)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": history,
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total_items":  total,
			"total_pages":  (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func parseUint(s string) (int, error) {
	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	return val, err
}

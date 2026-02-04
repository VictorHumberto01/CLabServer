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
	filterUserID := c.Query("user_id")
	filterClassroomID := c.Query("classroomId")

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

	// Preload Exercise as well since we might be showing exercise titles
	query := initializers.DB.Model(&models.History{}).Preload("User").Preload("Exercise")

	// If not admin/teacher, restrict to own history
	if u.Role != "ADMIN" && u.Role != "TEACHER" {
		query = query.Where("histories.user_id = ?", u.ID)
	} else {
		// Admin/Teacher can filter by specific user if provided
		if filterUserID != "" {
			query = query.Where("histories.user_id = ?", filterUserID)
		}
	}

	if filterClassroomID != "" {
		// Join with exercises to filter by classroom
		query = query.Joins("JOIN exercises ON exercises.id = histories.exercise_id").
			Where("exercises.classroom_id = ?", filterClassroomID)
	}

	if search != "" {
		query = query.Where("histories.code LIKE ?", "%"+search+"%")
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

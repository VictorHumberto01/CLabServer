package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vitub/CLabServer/internal/dtos"
	"github.com/vitub/CLabServer/internal/initializers"
	"github.com/vitub/CLabServer/internal/models"
)

func CreateTopic(c *gin.Context) {
	classroomId := c.Param("id")
	var req dtos.CreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: err.Error()})
		return
	}

	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var classroom models.Classroom
	if err := initializers.DB.First(&classroom, classroomId).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Classroom not found"})
		return
	}

	if classroom.TeacherID != currentUser.ID {
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{Error: "Not authorized"})
		return
	}

	topic := models.ExerciseTopic{
		ClassroomID: classroom.ID,
		Title:       req.Title,
	}

	if err := initializers.DB.Create(&topic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to create topic"})
		return
	}

	c.JSON(http.StatusCreated, dtos.SuccessResponse{
		Success: true,
		Data: dtos.TopicResponse{
			ID:          topic.ID,
			ClassroomID: topic.ClassroomID,
			Title:       topic.Title,
		},
	})
}

func ListTopics(c *gin.Context) {
	classroomId := c.Param("id")

	var topics []models.ExerciseTopic
	if err := initializers.DB.Preload("Exercises").Where("classroom_id = ?", classroomId).Find(&topics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to fetch topics"})
		return
	}

	var response []dtos.TopicResponse
	for _, t := range topics {
		var exercises []dtos.ExerciseResponse
		for _, ex := range t.Exercises {
			exercises = append(exercises, dtos.ExerciseResponse{
				ID:             ex.ID,
				ClassroomID:    ex.ClassroomID,
				TopicID:        ex.TopicID,
				Title:          ex.Title,
				Description:    ex.Description,
				ExpectedOutput: ex.ExpectedOutput,
				InitialCode:    ex.InitialCode,
				CreatedAt:      ex.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
		response = append(response, dtos.TopicResponse{
			ID:          t.ID,
			ClassroomID: t.ClassroomID,
			Title:       t.Title,
			Exercises:   exercises,
		})
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    response,
	})
}

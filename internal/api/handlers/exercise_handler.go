package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vitub/CLabServer/internal/dtos"
	"github.com/vitub/CLabServer/internal/initializers"
	"github.com/vitub/CLabServer/internal/models"
)

func CreateExercise(c *gin.Context) {
	classroomId := c.Param("id")
	var req dtos.CreateExerciseRequest
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
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{Error: "Not authorized to create exercises for this classroom"})
		return
	}

	exercise := models.Exercise{
		ClassroomID:    classroom.ID,
		TopicID:        req.TopicID,
		Title:          req.Title,
		Description:    req.Description,
		ExpectedOutput: req.ExpectedOutput,
		InitialCode:    req.InitialCode,
	}

	if err := initializers.DB.Create(&exercise).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to create exercise"})
		return
	}

	c.JSON(http.StatusCreated, dtos.SuccessResponse{
		Success: true,
		Data: dtos.ExerciseResponse{
			ID:             exercise.ID,
			ClassroomID:    exercise.ClassroomID,
			TopicID:        exercise.TopicID,
			Title:          exercise.Title,
			Description:    exercise.Description,
			ExpectedOutput: exercise.ExpectedOutput,
			InitialCode:    exercise.InitialCode,
			CreatedAt:      exercise.CreatedAt.Format(time.RFC3339),
		},
	})
}

func ListExercises(c *gin.Context) {
	classroomId := c.Param("id")

	var exercises []models.Exercise
	if err := initializers.DB.Where("classroom_id = ?", classroomId).Find(&exercises).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to fetch exercises"})
		return
	}

	var response []dtos.ExerciseResponse
	for _, ex := range exercises {
		response = append(response, dtos.ExerciseResponse{
			ID:             ex.ID,
			ClassroomID:    ex.ClassroomID,
			TopicID:        ex.TopicID,
			Title:          ex.Title,
			Description:    ex.Description,
			ExpectedOutput: ex.ExpectedOutput,
			InitialCode:    ex.InitialCode,
			CreatedAt:      ex.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    response,
	})
}

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vitub/CLabServer/internal/ai"
	"github.com/vitub/CLabServer/internal/dtos"
	"github.com/vitub/CLabServer/internal/models"
)

type GenerateQuestionsRequest struct {
	NumQuestions        int     `json:"numQuestions"`
	VariantsPerQuestion int     `json:"variantsPerQuestion"`
	Difficulty          string  `json:"difficulty"`
	Topic               string  `json:"topic"`
	NotePerQuestion     float64 `json:"notePerQuestion"`
}

func GenerateQuestions(c *gin.Context) {
	classroomId := c.Param("id")

	user, _ := c.Get("user")
	currentUser := user.(models.User)

	classroom, err := loadClassroomWithTeachers(classroomId)
	if err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Classroom not found"})
		return
	}

	if !isTeacherOfClassroom(currentUser.ID, classroom) {
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{Error: "Not authorized"})
		return
	}

	var req GenerateQuestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: err.Error()})
		return
	}

	// Defaults
	if req.NumQuestions <= 0 {
		req.NumQuestions = 3
	}
	if req.VariantsPerQuestion <= 0 {
		req.VariantsPerQuestion = 2
	}
	if req.Difficulty == "" {
		req.Difficulty = "médio"
	}
	if req.Topic == "" {
		req.Topic = "programação C geral"
	}
	if req.NotePerQuestion <= 0 {
		req.NotePerQuestion = 10.0
	}

	questions, err := ai.GenerateExamQuestions(
		req.NumQuestions,
		req.VariantsPerQuestion,
		req.Difficulty,
		req.Topic,
		req.NotePerQuestion,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Falha ao gerar questões: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    questions,
	})
}

package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vitub/CLabServer/internal/dtos"
	"github.com/vitub/CLabServer/internal/initializers"
	"github.com/vitub/CLabServer/internal/models"
)

func ListExams(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	folderId := c.Query("folderId")

	query := initializers.DB.Preload("Exercises").
		Where("teacher_id = ? AND is_exam = ?", currentUser.ID, true)

	if folderId == "none" {
		query = query.Where("folder_id IS NULL")
	} else if folderId != "" {
		query = query.Where("folder_id = ?", folderId)
	}

	var topics []models.ExerciseTopic
	query.Order("created_at desc").Find(&topics)

	var response []map[string]interface{}
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
				ExamMaxNote:    ex.ExamMaxNote,
				VariantGroupID: ex.VariantGroupID,
				CreatedAt:      ex.CreatedAt.Format(time.RFC3339),
			})
		}
		response = append(response, map[string]interface{}{
			"id":            t.ID,
			"classroomId":   t.ClassroomID,
			"teacherId":     t.TeacherID,
			"folderId":      t.FolderID,
			"title":         t.Title,
			"exercises":     exercises,
			"expireDate":    t.ExpireDate,
			"isExam":        t.IsExam,
			"createdAt":     t.CreatedAt.Format(time.RFC3339),
			"questionCount": countQuestionGroups(t.Exercises),
		})
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    response,
	})
}

func countQuestionGroups(exercises []models.Exercise) int {
	groups := make(map[string]bool)
	for _, ex := range exercises {
		if ex.VariantGroupID != "" {
			groups[ex.VariantGroupID] = true
		} else {
			groups[strconv.FormatUint(uint64(ex.ID), 10)] = true
		}
	}
	return len(groups)
}

func CreateExam(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var req struct {
		Title      string                            `json:"title" binding:"required"`
		ExpireDate *time.Time                        `json:"expireDate"`
		FolderID   *uint                             `json:"folderId"`
		Exercises  []dtos.CreateExerciseGroupRequest `json:"exercises"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: err.Error()})
		return
	}

	topic := models.ExerciseTopic{
		TeacherID:  currentUser.ID,
		FolderID:   req.FolderID,
		Title:      req.Title,
		ExpireDate: req.ExpireDate,
		IsExam:     true,
	}

	if err := initializers.DB.Create(&topic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to create exam"})
		return
	}

	for _, group := range req.Exercises {
		for _, variant := range group.Variants {
			exercise := models.Exercise{
				TopicID:        &topic.ID,
				Title:          variant.Title,
				Description:    variant.Description,
				ExpectedOutput: variant.ExpectedOutput,
				InitialCode:    variant.InitialCode,
				ExamMaxNote:    variant.ExamMaxNote,
				VariantGroupID: group.VariantGroupID,
			}
			initializers.DB.Create(&exercise)
		}
	}

	c.JSON(http.StatusCreated, dtos.SuccessResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":    topic.ID,
			"title": topic.Title,
		},
	})
}

func AssignExamToClassroom(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	examId := c.Param("id")

	var req struct {
		ClassroomID uint `json:"classroomId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: err.Error()})
		return
	}

	// Verify exam ownership
	var topic models.ExerciseTopic
	if err := initializers.DB.First(&topic, examId).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Exam not found"})
		return
	}
	if topic.TeacherID != currentUser.ID {
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{Error: "Not authorized"})
		return
	}

	// Verify classroom access
	classroom, err := loadClassroomWithTeachers(strconv.FormatUint(uint64(req.ClassroomID), 10))
	if err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Classroom not found"})
		return
	}
	if !isTeacherOfClassroom(currentUser.ID, classroom) {
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{Error: "Not a teacher of this classroom"})
		return
	}

	// Assign classroom to topic and all its exercises
	initializers.DB.Model(&topic).Update("classroom_id", req.ClassroomID)
	initializers.DB.Model(&models.Exercise{}).Where("topic_id = ?", topic.ID).Update("classroom_id", req.ClassroomID)

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Message: "Exam assigned to classroom",
	})
}

func DeleteExam(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	examId := c.Param("id")

	var topic models.ExerciseTopic
	if err := initializers.DB.First(&topic, examId).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Exam not found"})
		return
	}
	if topic.TeacherID != currentUser.ID {
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{Error: "Not authorized"})
		return
	}

	// Delete exercises first
	initializers.DB.Where("topic_id = ?", topic.ID).Delete(&models.Exercise{})
	initializers.DB.Delete(&topic)

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Message: "Exam deleted",
	})
}

func MoveExamToFolder(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	examId := c.Param("id")

	var req struct {
		FolderID *uint `json:"folderId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: err.Error()})
		return
	}

	var topic models.ExerciseTopic
	if err := initializers.DB.First(&topic, examId).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Exam not found"})
		return
	}
	if topic.TeacherID != currentUser.ID {
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{Error: "Not authorized"})
		return
	}

	initializers.DB.Model(&topic).Update("folder_id", req.FolderID)

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Message: "Exam moved",
	})
}

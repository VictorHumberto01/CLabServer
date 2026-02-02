package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vitub/CLabServer/internal/dtos"
	"github.com/vitub/CLabServer/internal/initializers"
	"github.com/vitub/CLabServer/internal/models"
)

func CreateClassroom(c *gin.Context) {
	var req dtos.CreateClassroomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: err.Error()})
		return
	}

	user, _ := c.Get("user")
	currentUser := user.(models.User)

	classroom := models.Classroom{
		Name:      req.Name,
		TeacherID: currentUser.ID,
	}

	if err := initializers.DB.Create(&classroom).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to create classroom"})
		return
	}

	c.JSON(http.StatusCreated, dtos.SuccessResponse{
		Success: true,
		Data: dtos.ClassroomResponse{
			ID:        classroom.ID,
			Name:      classroom.Name,
			TeacherID: classroom.TeacherID,
		},
	})
}

func ListClassrooms(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var classrooms []models.Classroom

	query := initializers.DB.Preload("Teacher").Model(&models.Classroom{})

	if currentUser.Role == models.RoleTeacher {
		query = query.Where("teacher_id = ?", currentUser.ID)
	} else if currentUser.Role == models.RoleUser {
		// Verify this logic: join with students table
		query = query.Joins("JOIN classroom_students cs ON cs.classroom_id = classrooms.id").
			Where("cs.student_id = ?", currentUser.ID)
	}

	if err := query.Find(&classrooms).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to fetch classrooms"})
		return
	}

	var response []dtos.ClassroomResponse
	for _, class := range classrooms {
		response = append(response, dtos.ClassroomResponse{
			ID:        class.ID,
			Name:      class.Name,
			TeacherID: class.TeacherID,
			Teacher: &dtos.UserResponse{
				ID:    class.Teacher.ID,
				Name:  class.Teacher.Name,
				Email: class.Teacher.Email,
				Role:  class.Teacher.Role,
			},
		})
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    response,
	})
}

func AddStudent(c *gin.Context) {
	classroomID := c.Param("id")
	var req dtos.AddStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: err.Error()})
		return
	}

	user, _ := c.Get("user")
	currentUser := user.(models.User)

	// 1. Find Classroom and verify ownership
	var classroom models.Classroom
	if err := initializers.DB.First(&classroom, classroomID).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Classroom not found"})
		return
	}

	if classroom.TeacherID != currentUser.ID {
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{Error: "Not authorized to manage this classroom"})
		return
	}

	// 2. Find Student by Email
	var student models.User
	if err := initializers.DB.Where("email = ?", req.Email).First(&student).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Student email not found"})
		return
	}

	// 3. Add Association
	if err := initializers.DB.Model(&classroom).Association("Students").Append(&student); err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to add student"})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Message: "Student added successfully",
	})
}

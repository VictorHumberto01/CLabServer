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

	query := initializers.DB.Preload("Teacher").Preload("Students").Model(&models.Classroom{})

	if currentUser.Role == models.RoleTeacher {
		query = query.Where("teacher_id = ?", currentUser.ID)
	} else if currentUser.Role == models.RoleUser {
		query = query.Joins("JOIN classroom_students cs ON cs.classroom_id = classrooms.id").
			Where("cs.user_id = ?", currentUser.ID)
	}

	if err := query.Find(&classrooms).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to fetch classrooms"})
		return
	}

	var response []dtos.ClassroomResponse
	for _, class := range classrooms {
		var totalExercises int64
		initializers.DB.Model(&models.Exercise{}).Where("classroom_id = ?", class.ID).Count(&totalExercises)

		response = append(response, dtos.ClassroomResponse{
			ID:           class.ID,
			Name:         class.Name,
			TeacherID:    class.TeacherID,
			ActiveExamID: class.ActiveExamTopicID,
			Teacher: &dtos.UserResponse{
				ID:    class.Teacher.ID,
				Name:  class.Teacher.Name,
				Email: class.Teacher.Email,
				Role:  class.Teacher.Role,
			},
			Students: func() []dtos.UserResponse {
				var students []dtos.UserResponse
				for _, s := range class.Students {
					var completedCount int64
					initializers.DB.Model(&models.History{}).
						Joins("JOIN exercises e ON e.id = histories.exercise_id").
						Where("histories.user_id = ? AND histories.is_success = ? AND e.classroom_id = ?", s.ID, true, class.ID).
						Distinct("histories.exercise_id").
						Count(&completedCount)

					students = append(students, dtos.UserResponse{
						ID:                 s.ID,
						Name:               s.Name,
						Email:              s.Email,
						Role:               s.Role,
						CompletedExercises: int(completedCount),
						TotalExercises:     int(totalExercises),
					})
				}
				return students
			}(),
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

	var classroom models.Classroom
	if err := initializers.DB.First(&classroom, classroomID).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Classroom not found"})
		return
	}

	if classroom.TeacherID != currentUser.ID {
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{Error: "Not authorized to manage this classroom"})
		return
	}

	var student models.User
	if req.Email != "" {
		if err := initializers.DB.Where("email = ?", req.Email).First(&student).Error; err != nil {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Student email not found"})
			return
		}
	} else if req.Matricula != "" {
		if err := initializers.DB.Where("matricula = ?", req.Matricula).First(&student).Error; err != nil {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Student matricula not found"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "Email or matricula is required"})
		return
	}

	if err := initializers.DB.Model(&classroom).Association("Students").Append(&student); err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to add student"})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Message: "Student added successfully",
	})
}

func DeleteClassroom(c *gin.Context) {
	classroomID := c.Param("id")
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var classroom models.Classroom
	if err := initializers.DB.First(&classroom, classroomID).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Classroom not found"})
		return
	}

	if classroom.TeacherID != currentUser.ID {
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{Error: "Not authorized to delete this classroom"})
		return
	}

	if err := initializers.DB.Delete(&classroom).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to delete classroom"})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Message: "Classroom deleted successfully",
	})
}
func RemoveStudent(c *gin.Context) {
	classroomID := c.Param("id")
	studentID := c.Param("studentId")

	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var classroom models.Classroom
	if err := initializers.DB.First(&classroom, classroomID).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Classroom not found"})
		return
	}

	if classroom.TeacherID != currentUser.ID {
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{Error: "Not authorized to manage this classroom"})
		return
	}

	var student models.User
	if err := initializers.DB.First(&student, studentID).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Student not found"})
		return
	}

	if err := initializers.DB.Model(&classroom).Association("Students").Delete(&student); err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to remove student from classroom"})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Message: "Student removed from classroom successfully",
	})
}

func ToggleExamMode(c *gin.Context) {
	classroomID := c.Param("id")
	var req dtos.UpdateClassroomExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: err.Error()})
		return
	}

	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var classroom models.Classroom
	if err := initializers.DB.First(&classroom, classroomID).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Classroom not found"})
		return
	}

	if classroom.TeacherID != currentUser.ID {
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{Error: "Not authorized to manage this classroom"})
		return
	}

	if req.ActiveExamID != nil {
		var topic models.ExerciseTopic
		if err := initializers.DB.First(&topic, *req.ActiveExamID).Error; err != nil {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "Exam topic not found"})
			return
		}
		if topic.ClassroomID != classroom.ID {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "Topic does not belong to this classroom"})
			return
		}
		if !topic.IsExam {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "Selected topic is not marked as an exam"})
			return
		}
	}

	if err := initializers.DB.Model(&classroom).Update("active_exam_topic_id", req.ActiveExamID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to update exam mode"})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Message: "Exam mode updated successfully",
	})
}

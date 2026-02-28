package handlers

import (
	"fmt"
	"hash/fnv"
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

	classroom, err := loadClassroomWithTeachers(classroomId)
	if err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Classroom not found"})
		return
	}

	if !isTeacherOfClassroom(currentUser.ID, classroom) {
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{Error: "Not authorized"})
		return
	}

	topic := models.ExerciseTopic{
		ClassroomID: &classroom.ID,
		TeacherID:   currentUser.ID,
		Title:       req.Title,
		ExpireDate:  req.ExpireDate,
		IsExam:      req.IsExam,
	}

	if err := initializers.DB.Create(&topic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to create topic"})
		return
	}

	// Create nested exercises if any
	for _, group := range req.Exercises {
		for _, variant := range group.Variants {
			exercise := models.Exercise{
				ClassroomID:    &classroom.ID,
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
		Data: dtos.TopicResponse{
			ID:          topic.ID,
			ClassroomID: topic.ClassroomID,
			Title:       topic.Title,
			ExpireDate:  topic.ExpireDate,
			IsExam:      topic.IsExam,
		},
	})
}

func ListTopics(c *gin.Context) {
	classroomId := c.Param("id")
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var topics []models.ExerciseTopic
	if err := initializers.DB.Preload("Exercises").Where("classroom_id = ?", classroomId).Find(&topics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to fetch topics"})
		return
	}

	// Filter variants for students
	if currentUser.Role == models.RoleUser {
		for i := range topics {
			if topics[i].IsExam && len(topics[i].Exercises) > 0 {
				grouped := make(map[string][]models.Exercise)
				var noGroup []models.Exercise

				for _, ex := range topics[i].Exercises {
					if ex.VariantGroupID != "" {
						grouped[ex.VariantGroupID] = append(grouped[ex.VariantGroupID], ex)
					} else {
						noGroup = append(noGroup, ex)
					}
				}

				var filteredExercises []models.Exercise
				filteredExercises = append(filteredExercises, noGroup...)

				for groupId, groupVars := range grouped {
					if len(groupVars) == 1 {
						filteredExercises = append(filteredExercises, groupVars[0])
					} else if len(groupVars) > 1 {
						// Deterministic hash based on Student ID, Topic ID, and Variant Group ID
						hashInput := fmt.Sprintf("%d-%d-%s", currentUser.ID, topics[i].ID, groupId)
						h := fnv.New32a()
						h.Write([]byte(hashInput))
						selectedIndex := h.Sum32() % uint32(len(groupVars))
						filteredExercises = append(filteredExercises, groupVars[selectedIndex])
					}
				}
				topics[i].Exercises = filteredExercises
			}
		}
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
				ExamMaxNote:    ex.ExamMaxNote,
				VariantGroupID: ex.VariantGroupID,
				CreatedAt:      ex.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
		response = append(response, dtos.TopicResponse{
			ID:          t.ID,
			ClassroomID: t.ClassroomID,
			Title:       t.Title,
			Exercises:   exercises,
			ExpireDate:  t.ExpireDate,
			IsExam:      t.IsExam,
		})
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    response,
	})
}

func DeleteTopic(c *gin.Context) {
	classroomId := c.Param("id")
	topicId := c.Param("topicId")

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

	var topic models.ExerciseTopic
	if err := initializers.DB.Where("id = ? AND classroom_id = ?", topicId, classroomId).First(&topic).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Topic not found"})
		return
	}

	if err := initializers.DB.Delete(&topic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to delete topic"})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Message: "Topic deleted successfully",
	})
}

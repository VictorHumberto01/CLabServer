package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vitub/CLabServer/internal/dtos"
	"github.com/vitub/CLabServer/internal/initializers"
	"github.com/vitub/CLabServer/internal/models"
)

func ListFolders(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var folders []models.ExamFolder
	initializers.DB.Where("teacher_id = ?", currentUser.ID).Order("created_at desc").Find(&folders)

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    folders,
	})
}

func CreateFolder(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: err.Error()})
		return
	}

	folder := models.ExamFolder{
		Name:      req.Name,
		TeacherID: currentUser.ID,
	}

	if err := initializers.DB.Create(&folder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to create folder"})
		return
	}

	c.JSON(http.StatusCreated, dtos.SuccessResponse{
		Success: true,
		Data:    folder,
	})
}

func RenameFolder(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	folderId := c.Param("id")

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: err.Error()})
		return
	}

	var folder models.ExamFolder
	if err := initializers.DB.Where("id = ? AND teacher_id = ?", folderId, currentUser.ID).First(&folder).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Folder not found"})
		return
	}

	folder.Name = req.Name
	initializers.DB.Save(&folder)

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    folder,
	})
}

func DeleteFolder(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	folderId := c.Param("id")

	var folder models.ExamFolder
	if err := initializers.DB.Where("id = ? AND teacher_id = ?", folderId, currentUser.ID).First(&folder).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Folder not found"})
		return
	}

	// Unlink exams from folder (don't delete them)
	initializers.DB.Model(&models.ExerciseTopic{}).Where("folder_id = ?", folder.ID).Update("folder_id", nil)

	initializers.DB.Delete(&folder)

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Message: "Folder deleted",
	})
}

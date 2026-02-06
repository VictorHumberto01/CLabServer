package handlers

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/vitub/CLabServer/internal/dtos"
	"github.com/vitub/CLabServer/internal/initializers"
	"github.com/vitub/CLabServer/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func LoginWithToken(c *gin.Context) {
	var req dtos.LoginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Invalid input: " + err.Error(),
		})
		return
	}

	var users models.User
	result := initializers.DB.Where("email = ?", req.Email).First(&users)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
				Success: false,
				Error:   "Invalid email or password",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Something went wrong. We are looking into it.",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(users.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Success: false,
			Error:   "Invalid email or password",
		})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": users.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    dtos.TokenResponse{Token: tokenString},
	})

}

func LoginMatricula(c *gin.Context) {
	var req dtos.LoginMatriculaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Matrícula e senha são obrigatórios",
		})
		return
	}

	var user models.User
	result := initializers.DB.Where("matricula = ?", req.Matricula).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
				Success: false,
				Error:   "Matrícula ou senha inválida",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Erro interno",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Success: false,
			Error:   "Matrícula ou senha inválida",
		})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Erro ao gerar token",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data: gin.H{
			"token":      tokenString,
			"needsSetup": user.Email == "",
		},
	})
}

func LoginWithCookie(c *gin.Context) {
	var req dtos.LoginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Invalid input: " + err.Error(),
		})
		return
	}

	var users models.User
	result := initializers.DB.Where("email = ?", req.Email).First(&users)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
				Success: false,
				Error:   "Invalid email or password",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Something went wrong. We are looking into it.",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(users.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Success: false,
			Error:   "Invalid email or password",
		})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": users.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to generate token",
		})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"Authorization",
		tokenString,
		3600*24*30,
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    dtos.TokenResponse{Token: tokenString},
	})

}

func Validate(c *gin.Context) {
	user, _ := c.Get("user")

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    user,
	})
}

func UpdateProfile(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Success: false,
			Error:   "Não autorizado",
		})
		return
	}
	user := currentUser.(models.User)

	var req dtos.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Dados inválidos: " + err.Error(),
		})
		return
	}

	updates := map[string]interface{}{}

	if req.Email != "" {
		updates["email"] = req.Email
	}

	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Success: false,
				Error:   "Erro ao processar senha",
			})
			return
		}
		updates["password"] = string(hash)
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Nenhum dado para atualizar",
		})
		return
	}

	if err := initializers.DB.Model(&user).Updates(updates).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			c.JSON(http.StatusConflict, dtos.ErrorResponse{
				Success: false,
				Error:   "Email já em uso",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Erro ao atualizar perfil",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Message: "Perfil atualizado com sucesso",
	})
}

func CreateUser(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}
	creator := currentUser.(models.User)

	var req dtos.CreateUserByRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Invalid input: " + err.Error(),
		})
		return
	}

	targetRole := models.RoleUser
	if req.Role != "" {
		targetRole = req.Role
	}

	switch creator.Role {
	case models.RoleAdmin:

		if targetRole != models.RoleTeacher && targetRole != models.RoleUser {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Success: false,
				Error:   "Invalid role. Admins can create TEACHER or USER",
			})
			return
		}
	case models.RoleTeacher:
		if targetRole != models.RoleUser {
			c.JSON(http.StatusForbidden, dtos.ErrorResponse{
				Success: false,
				Error:   "Teachers can only create students (USER role)",
			})
			return
		}
	default:
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{
			Success: false,
			Error:   "You don't have permission to create users",
		})
		return
	}

	// Generate random password if not provided
	initialPassword := req.Password
	if initialPassword == "" {
		initialPassword = generateSimplePassword()
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(initialPassword), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to hash password",
		})
		return
	}

	user := models.User{
		Name:      req.Name,
		Email:     req.Email,
		Matricula: req.Matricula,
		Password:  string(hash),
		Role:      targetRole,
	}

	if err := initializers.DB.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			c.JSON(http.StatusConflict, dtos.ErrorResponse{
				Success: false,
				Error:   "Email ou matrícula já em uso",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to create user: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dtos.SuccessResponse{
		Success: true,
		Data: dtos.CreateUserResponse{
			ID:              user.ID,
			Name:            user.Name,
			Email:           user.Email,
			Matricula:       user.Matricula,
			Role:            user.Role,
			InitialPassword: initialPassword,
		},
	})
}

func generateSimplePassword() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 6)
	for i := range result {
		result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
		time.Sleep(time.Nanosecond)
	}
	return string(result)
}

func ListUsers(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}
	requester := currentUser.(models.User)

	if requester.Role != models.RoleAdmin && requester.Role != models.RoleTeacher {
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{
			Success: false,
			Error:   "You don't have permission to list users",
		})
		return
	}

	roleFilter := c.Query("role")
	classroomID := c.Query("classroomId")

	var users []models.User
	query := initializers.DB.Model(&models.User{})

	if roleFilter != "" {
		query = query.Where("role = ?", roleFilter)
	}
	if requester.Role == models.RoleTeacher {
		query = query.Where("role = ?", models.RoleUser)
	}

	if classroomID != "" {
		query = query.Joins("JOIN classroom_students ON classroom_students.user_id = users.id").
			Where("classroom_students.classroom_id = ?", classroomID)
	}

	var total int64
	query.Count(&total)

	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch users",
		})
		return
	}

	var userResponses []dtos.UserResponse
	for _, u := range users {
		userResponses = append(userResponses, dtos.UserResponse{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			Matricula: u.Matricula,
			Role:      u.Role,
		})
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data: dtos.UsersListResponse{
			Users: userResponses,
			Total: total,
		},
	})
}

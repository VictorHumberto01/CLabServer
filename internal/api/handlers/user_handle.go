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

func SignUp(c *gin.Context) {
	var req dtos.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Invalid input: " + err.Error(),
		})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to hash password",
		})
		return
	}

	user := models.User{
		Email:    req.Email,
		Password: string(hash),
		Role:     models.RoleUser,
	}

	if err := initializers.DB.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			c.JSON(http.StatusConflict, dtos.ErrorResponse{
				Success: false,
				Error:   "Email already in use",
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
		Data:    dtos.IDResponse{ID: user.ID},
	})
}

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

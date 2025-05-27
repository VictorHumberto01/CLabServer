package handlers

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"

	"clab-server/internal/api/middleware"
	"clab-server/internal/database/models"
)

type AuthHandler struct {
	db             *gorm.DB
	authMiddleware *middleware.AuthMiddleware
}

func NewAuthHandler(db *gorm.DB, authMiddleware *middleware.AuthMiddleware) *AuthHandler {
	return &AuthHandler{
		db:             db,
		authMiddleware: authMiddleware,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !user.CheckPassword(req.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.authMiddleware.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate role
	role := models.UserRole(req.Role)
	if role != models.RoleTeacher && role != models.RoleStudent {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	user := models.User{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Role:     role,
	}

	if err := user.HashPassword(); err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	if err := h.db.Create(&user).Error; err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	token, err := h.authMiddleware.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(AuthResponse{
		Token: token,
		User:  user,
	})
}

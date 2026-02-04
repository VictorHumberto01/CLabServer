package dtos

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID                 uint   `json:"id"`
	Name               string `json:"name"`
	Email              string `json:"email"`
	Role               string `json:"role"`
	CompletedExercises int    `json:"completedExercises,omitempty"`
	TotalExercises     int    `json:"totalExercises,omitempty"`
}

// CreateUserByRoleRequest is used by admins/teachers to create users
type CreateUserByRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	Role        string `json:"role"` // Optional: TEACHER or USER (default: USER)
	ClassroomID *uint  `json:"classroomId,omitempty"`
}

// UsersListResponse for listing users
type UsersListResponse struct {
	Users []UserResponse `json:"users"`
	Total int64          `json:"total"`
}

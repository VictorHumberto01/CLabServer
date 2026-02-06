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

type LoginMatriculaRequest struct {
	Matricula string `json:"matricula" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"omitempty,min=4"`
}

type UserResponse struct {
	ID                 uint   `json:"id"`
	Name               string `json:"name"`
	Email              string `json:"email"`
	Matricula          string `json:"matricula"`
	Role               string `json:"role"`
	NeedsSetup         bool   `json:"needsSetup"`
	CompletedExercises int    `json:"completedExercises,omitempty"`
	TotalExercises     int    `json:"totalExercises,omitempty"`
}

type CreateUserResponse struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Matricula       string `json:"matricula"`
	Role            string `json:"role"`
	InitialPassword string `json:"initialPassword,omitempty"`
}

type CreateUserByRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email"`
	Matricula   string `json:"matricula"`
	Password    string `json:"password"`
	Role        string `json:"role"`
	ClassroomID *uint  `json:"classroomId,omitempty"`
}

// UsersListResponse for listing users
type UsersListResponse struct {
	Users []UserResponse `json:"users"`
	Total int64          `json:"total"`
}

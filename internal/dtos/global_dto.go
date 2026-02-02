package dtos

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type IDResponse struct {
	ID uint `json:"id"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

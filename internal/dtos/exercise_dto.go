package dtos

import (
	"time"
)

type CreateExerciseRequest struct {
	TopicID        *uint   `json:"topicId"`
	Title          string  `json:"title" binding:"required"`
	Description    string  `json:"description" binding:"required"`
	ExpectedOutput string  `json:"expectedOutput"`
	InitialCode    string  `json:"initialCode"`
	ExamMaxNote    float64 `json:"examMaxNote"`
}

type ExerciseResponse struct {
	ID             uint    `json:"id"`
	ClassroomID    uint    `json:"classroomId"`
	TopicID        *uint   `json:"topicId"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	ExpectedOutput string  `json:"expectedOutput"`
	InitialCode    string  `json:"initialCode"`
	CreatedAt      string  `json:"createdAt"`
	ExamMaxNote    float64 `json:"examMaxNote"`
}

type CreateTopicRequest struct {
	Title      string     `json:"title" binding:"required"`
	ExpireDate *time.Time `json:"expireDate"`
	IsExam     bool       `json:"isExam"`
}

type TopicResponse struct {
	ID          uint               `json:"id"`
	ClassroomID uint               `json:"classroomId"`
	Title       string             `json:"title"`
	Exercises   []ExerciseResponse `json:"exercises,omitempty"`
	ExpireDate  *time.Time         `json:"expireDate"`
	IsExam      bool               `json:"isExam"`
}

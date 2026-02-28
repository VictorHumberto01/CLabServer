package models

import (
	"gorm.io/gorm"
)

type Exercise struct {
	gorm.Model
	ClassroomID    *uint          `json:"classroomId"`
	Classroom      *Classroom     `json:"classroom,omitempty" gorm:"foreignKey:ClassroomID"`
	TopicID        *uint          `json:"topicId"`
	Topic          *ExerciseTopic `json:"topic,omitempty" gorm:"foreignKey:TopicID"`
	Title          string         `json:"title" gorm:"not null"`
	Description    string         `json:"description"`
	ExpectedOutput string         `json:"expectedOutput"`
	InitialCode    string         `json:"initialCode"`
	ExamMaxNote    float64        `json:"examMaxNote" gorm:"default:10.0"`
	VariantGroupID string         `json:"variantGroupId"`
}

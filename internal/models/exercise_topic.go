package models

import (
	"time"

	"gorm.io/gorm"
)

type ExerciseTopic struct {
	gorm.Model
	ClassroomID uint       `json:"classroomId" gorm:"not null"`
	Classroom   Classroom  `json:"classroom,omitempty" gorm:"foreignKey:ClassroomID"`
	Title       string     `json:"title" gorm:"not null"`
	Exercises   []Exercise `json:"exercises,omitempty" gorm:"foreignKey:TopicID"`
	ExpireDate  *time.Time `json:"expireDate"`
	IsExam      bool       `json:"isExam" gorm:"default:false"`
}

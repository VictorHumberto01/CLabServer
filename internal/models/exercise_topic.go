package models

import "gorm.io/gorm"

type ExerciseTopic struct {
	gorm.Model
	ClassroomID uint       `json:"classroomId" gorm:"not null"`
	Classroom   Classroom  `json:"classroom,omitempty" gorm:"foreignKey:ClassroomID"`
	Title       string     `json:"title" gorm:"not null"`
	Exercises   []Exercise `json:"exercises,omitempty" gorm:"foreignKey:TopicID"`
}

package models

import (
	"time"

	"gorm.io/gorm"
)

type ExerciseTopic struct {
	gorm.Model
	ClassroomID *uint       `json:"classroomId"`
	Classroom   *Classroom  `json:"classroom,omitempty" gorm:"foreignKey:ClassroomID"`
	TeacherID   uint        `json:"teacherId" gorm:"default:0"`
	FolderID    *uint       `json:"folderId"`
	Folder      *ExamFolder `json:"folder,omitempty" gorm:"foreignKey:FolderID"`
	Title       string      `json:"title" gorm:"not null"`
	Exercises   []Exercise  `json:"exercises,omitempty" gorm:"foreignKey:TopicID"`
	ExpireDate  *time.Time  `json:"expireDate"`
	IsExam      bool        `json:"isExam" gorm:"default:false"`
}

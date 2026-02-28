package models

import "gorm.io/gorm"

type ExamFolder struct {
	gorm.Model
	Name      string `json:"name" gorm:"not null"`
	TeacherID uint   `json:"teacherId" gorm:"not null"`
	Teacher   User   `json:"teacher,omitempty" gorm:"foreignKey:TeacherID"`
}

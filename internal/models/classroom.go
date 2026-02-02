package models

import "gorm.io/gorm"

type Classroom struct {
	gorm.Model
	Name      string `json:"name" gorm:"not null"`
	TeacherID uint   `json:"teacherId"`
	Teacher   User   `json:"teacher,omitempty" gorm:"foreignKey:TeacherID"`
	Students  []User `json:"students,omitempty" gorm:"many2many:classroom_students;"`
}

// ClassroomStudent is used for the many-to-many relationship join table.
// It can be extended with fields like 'Grade' or 'EnrolledAt' in the future.
type ClassroomStudent struct {
	ClassroomID uint
	StudentID   uint
}

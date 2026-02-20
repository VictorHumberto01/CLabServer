package models

import "gorm.io/gorm"

type Classroom struct {
	gorm.Model
	Name              string `json:"name" gorm:"not null"`
	TeacherID         uint   `json:"teacherId"` // Owner of the classroom
	ActiveExamTopicID *uint  `json:"activeExamTopicId"`
	Teacher           User   `json:"teacher,omitempty" gorm:"foreignKey:TeacherID"`           // Owner
	Teachers          []User `json:"teachers,omitempty" gorm:"many2many:classroom_teachers;"` // Co-teachers
	Students          []User `json:"students,omitempty" gorm:"many2many:classroom_students;"`
}

type ClassroomStudent struct {
	ClassroomID uint
	StudentID   uint
}

type ClassroomTeacher struct {
	ClassroomID uint
	TeacherID   uint
}

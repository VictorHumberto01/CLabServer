package models

import "gorm.io/gorm"

const (
	RoleUser    = "USER"
	RoleAdmin   = "ADMIN"
	RoleTeacher = "TEACHER"
)

type User struct {
	gorm.Model
	ID       uint
	Name     string `gorm:"not null"`
	Email    string `gorm:"unique;index;not null"`
	Password string `gorm:"not null" json:"-"`
	Role     string `gorm:"default:USER;not null"`
	History  []History
}

func (u *User) isAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) isTeacher() bool {
	return u.Role == RoleTeacher
}

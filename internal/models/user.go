package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	RoleUser    = "USER"
	RoleAdmin   = "ADMIN"
	RoleTeacher = "TEACHER"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Name      string         `gorm:"not null" json:"name"`
	Email     string         `gorm:"unique;index;not null" json:"email"`
	Password  string         `gorm:"not null" json:"-"`
	Role      string         `gorm:"default:USER;not null" json:"role"`
	History   []History      `json:"history"`
}

func (u *User) isAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) isTeacher() bool {
	return u.Role == RoleTeacher
}

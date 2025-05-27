package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleTeacher  UserRole = "teacher"
	RoleStudent  UserRole = "student"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"-" gorm:"not null"` // "-" means this field won't be included in JSON
	Name      string    `json:"name" gorm:"not null"`
	Role      UserRole  `json:"role" gorm:"type:varchar(20);not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// HashPassword hashes the user's password
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword checks if the provided password matches the hashed password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
} 
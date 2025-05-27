package models

import (
	"time"
)

type Room struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Code        string    `json:"code" gorm:"unique;not null"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	TeacherID   uint      `json:"teacher_id" gorm:"not null"`
	Teacher     User      `json:"teacher" gorm:"foreignKey:TeacherID"`
	Tasks       []Task    `json:"tasks" gorm:"foreignKey:RoomID"`
	Students    []User    `json:"students" gorm:"many2many:room_students;"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Task struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	RoomID      uint         `json:"room_id" gorm:"not null"`
	Room        Room         `json:"room" gorm:"foreignKey:RoomID"`
	Title       string       `json:"title" gorm:"not null"`
	Description string       `json:"description"`
	TestCases   []TestCase   `json:"test_cases" gorm:"foreignKey:TaskID"`
	Submissions []Submission `json:"submissions" gorm:"foreignKey:TaskID"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type TestCase struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	TaskID         uint      `json:"task_id" gorm:"not null"`
	Input          string    `json:"input" gorm:"not null"`
	ExpectedOutput string    `json:"expected_output" gorm:"not null"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Submission struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TaskID    uint      `json:"task_id" gorm:"not null"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	Code      string    `json:"code" gorm:"type:text;not null"`
	Status    string    `json:"status" gorm:"type:varchar(20);not null"` // pending, compiling, running, completed, error
	Feedback  string    `json:"feedback" gorm:"type:text"`
	Score     float64   `json:"score"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

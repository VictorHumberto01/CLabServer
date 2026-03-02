package models

import (
	"time"

	"gorm.io/gorm"
)

type History struct {
	ID             uint      `gorm:"primarykey"`
	CreatedAt      time.Time `gorm:"index"`
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	UserID         uint           `gorm:"index"`
	User           User           `json:"user" gorm:"foreignKey:UserID"`
	ExerciseID     *uint          `json:"exerciseId,omitempty" gorm:"index"`
	Exercise       *Exercise      `json:"exercise,omitempty" gorm:"foreignKey:ExerciseID"`
	Code           string         `json:"code"`
	Input          string         `json:"input"`
	Output         string         `json:"output"`
	Error          string         `json:"error"`
	AIAnalysis     string         `json:"aiAnalysis"`
	TeacherGrading string         `json:"teacherGrading"`
	Score          float64        `json:"score"`
	IsSuccess      bool           `json:"isSuccess"`
}

package models

import "gorm.io/gorm"

type History struct {
	gorm.Model
	UserID         uint
	User           User      `json:"user" gorm:"foreignKey:UserID"`
	ExerciseID     *uint     `json:"exerciseId,omitempty"`
	Exercise       *Exercise `json:"exercise,omitempty" gorm:"foreignKey:ExerciseID"`
	Code           string    `json:"code"`
	Input          string    `json:"input"`
	Output         string    `json:"output"`
	Error          string    `json:"error"`
	AIAnalysis     string    `json:"aiAnalysis"`
	TeacherGrading string    `json:"teacherGrading"`
	Score          float64   `json:"score"`
	IsSuccess      bool      `json:"isSuccess"`
}

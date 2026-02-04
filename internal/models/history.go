package models

import "gorm.io/gorm"

type History struct {
	gorm.Model
	UserID     uint
	User       User      `json:"user" gorm:"foreignKey:UserID"`
	ExerciseID *uint     `json:"exerciseId,omitempty"`
	Exercise   *Exercise `json:"exercise,omitempty" gorm:"foreignKey:ExerciseID"`
	Code       string
	Input      string
	Output     string
	Error      string
	AIAnalysis string `json:"aiAnalysis"`
	IsSuccess  bool   `json:"isSuccess"`
}

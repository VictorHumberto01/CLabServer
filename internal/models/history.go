package models

import "gorm.io/gorm"

type History struct {
	gorm.Model
	UserID uint
	Code   string
	Input  string
	Output string
	Error  string
}

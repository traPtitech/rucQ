package model

import "gorm.io/gorm"

type Option struct {
	gorm.Model
	QuestionID uint
	Content    string
}

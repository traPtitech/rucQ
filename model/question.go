package model

import "gorm.io/gorm"

type Question struct {
	gorm.Model
	QuestionGroupID uint
	Title           string
	Description     string
	Type            string
	IsPublic        bool
	IsOpen          bool
	Options         []Option
	Answers         []Answer
}

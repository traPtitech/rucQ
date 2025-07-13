package model

import "gorm.io/gorm"

type QuestionType string

const (
	FreeTextQuestion       QuestionType = "free_text"
	FreeNumberQuestion     QuestionType = "free_number"
	SingleChoiceQuestion   QuestionType = "single"
	MultipleChoiceQuestion QuestionType = "multiple"
)

type Question struct {
	gorm.Model
	Type            QuestionType `gorm:"type:enum('free_text', 'free_number', 'single', 'multiple')"`
	QuestionGroupID uint
	Title           string
	Description     *string
	IsPublic        bool
	IsOpen          bool
	IsRequired      bool
	Options         []Option

	Answers []Answer
}

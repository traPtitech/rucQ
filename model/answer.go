package model

import "gorm.io/gorm"

type AnswerType string

const (
	FreeTextAnswer       AnswerType = "free_text"
	FreeNumberAnswer     AnswerType = "free_number"
	SingleChoiceAnswer   AnswerType = "single"
	MultipleChoiceAnswer AnswerType = "multiple"
)

type Answer struct {
	gorm.Model
	QuestionID        uint       `gorm:"uniqueIndex:idx_question_id_user_id"`
	UserID            string     `gorm:"uniqueIndex:idx_question_id_user_id"`
	Type              AnswerType `gorm:"type:enum('free_text', 'free_number', 'single', 'multiple')"`
	FreeTextContent   *string
	FreeNumberContent *float64
	SelectedOptions   []Option `gorm:"many2many:answer_options;ForeignKey:id;References:id"`
}

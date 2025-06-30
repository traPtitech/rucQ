//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package repository

import "github.com/traPtitech/rucQ/model"

type QuestionRepository interface {
	CreateQuestion(question *model.Question) error
	GetQuestions() ([]model.Question, error)
	GetQuestionByID(id uint) (*model.Question, error)
	DeleteQuestionByID(id uint) error
	UpdateQuestion(questionID uint, question *model.Question) error
}

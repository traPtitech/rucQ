//go:generate go tool mockgen -source=$GOFILE -destination=mockrepository/$GOFILE -package=mockrepository
package repository

import (
	"context"

	"github.com/traPtitech/rucQ/model"
)

type QuestionRepository interface {
	CreateQuestion(question *model.Question) error
	GetQuestions() ([]model.Question, error)
	GetQuestionByID(id uint) (*model.Question, error)
	GetQuestionCampID(ctx context.Context, questionID uint) (uint, error)
	DeleteQuestionByID(id uint) error
	UpdateQuestion(ctx context.Context, questionID uint, question *model.Question) error
}

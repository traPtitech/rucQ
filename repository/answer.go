//go:generate go tool mockgen -source=$GOFILE -destination=mockrepository/$GOFILE -package=mockrepository
package repository

import (
	"context"

	"github.com/traPtitech/rucQ/model"
)

type GetAnswersQuery struct {
	UserID                 *string
	QuestionGroupID        *uint
	QuestionID             *uint
	IncludePrivateAnswers  bool
	IncludeNonParticipants bool
}

type AnswerRepository interface {
	CreateAnswer(ctx context.Context, answer *model.Answer) error
	CreateAnswers(ctx context.Context, answers *[]model.Answer) error
	GetAnswerByID(ctx context.Context, id uint) (*model.Answer, error)
	GetAnswers(ctx context.Context, query GetAnswersQuery) ([]model.Answer, error)
	UpdateAnswer(ctx context.Context, answerID uint, answer *model.Answer) error
}

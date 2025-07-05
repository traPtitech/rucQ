//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package repository

import (
	"context"

	"github.com/traPtitech/rucQ/model"
)

type AnswerRepository interface {
	CreateAnswers(ctx context.Context, answers *[]model.Answer) error
	GetAnswerByID(ctx context.Context, id uint) (*model.Answer, error)
	GetAnswersByUserAndQuestionGroup(
		ctx context.Context,
		userID string,
		questionGroupID uint,
	) ([]model.Answer, error)
	UpdateAnswer(ctx context.Context, answerID uint, answer *model.Answer) error
}

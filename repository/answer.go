//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package repository

import (
	"context"

	"github.com/traPtitech/rucQ/model"
)

type AnswerRepository interface {
	CreateAnswers(ctx context.Context, answers *[]model.Answer) error
	GetAnswerByID(id uint) (*model.Answer, error)
	UpdateAnswer(answer *model.Answer) error
}

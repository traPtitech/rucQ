//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package repository

import (
	"context"

	"github.com/traPtitech/rucQ/model"
)

type QuestionGroupRepository interface {
	CreateQuestionGroup(questionGroup *model.QuestionGroup) error
	GetQuestionGroups(ctx context.Context, campID uint) ([]model.QuestionGroup, error)
	GetQuestionGroup(ID uint) (*model.QuestionGroup, error)
	UpdateQuestionGroup(ID uint, questionGroup *model.QuestionGroup) error
	DeleteQuestionGroup(ID uint) error
}

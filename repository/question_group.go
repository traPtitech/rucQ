//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package repository

import (
	"context"

	"github.com/traPtitech/rucQ/model"
)

type QuestionGroupRepository interface {
	CreateQuestionGroup(questionGroup *model.QuestionGroup) error
	GetQuestionGroups(ctx context.Context, campID uint) ([]model.QuestionGroup, error)
	GetQuestionGroup(ctx context.Context, ID uint) (*model.QuestionGroup, error)
	UpdateQuestionGroup(
		ctx context.Context,
		questionGroupID uint,
		questionGroup model.QuestionGroup,
	) error
	DeleteQuestionGroup(ID uint) error
}

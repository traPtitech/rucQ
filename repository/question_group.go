//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package repository

import "github.com/traP-jp/rucQ/backend/model"

type QuestionGroupRepository interface {
	CreateQuestionGroup(questionGroup *model.QuestionGroup) error
	GetQuestionGroups() ([]model.QuestionGroup, error)
	GetQuestionGroup(ID uint) (*model.QuestionGroup, error)
	UpdateQuestionGroup(ID uint, questionGroup *model.QuestionGroup) error
	DeleteQuestionGroup(ID uint) error
}

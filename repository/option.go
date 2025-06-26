//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package repository

import "github.com/traP-jp/rucQ/backend/model"

type GetOptionsQuery struct {
	QuestionID *uint
}

type OptionRepository interface {
	CreateOption(option *model.Option) error
	GetOptions(query *GetOptionsQuery) ([]model.Option, error)
}

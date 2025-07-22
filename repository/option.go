//go:generate go tool mockgen -source=$GOFILE -destination=mockrepository/$GOFILE -package=mockrepository
package repository

import "github.com/traPtitech/rucQ/model"

type GetOptionsQuery struct {
	QuestionID *uint
}

type OptionRepository interface {
	CreateOption(option *model.Option) error
	GetOptions(query *GetOptionsQuery) ([]model.Option, error)
}

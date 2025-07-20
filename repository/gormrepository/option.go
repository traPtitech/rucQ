package gormrepository

import (
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
)

func (r *Repository) CreateOption(option *model.Option) error {
	if err := r.db.Create(option).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetOptions(query *repository.GetOptionsQuery) ([]model.Option, error) {
	tx := r.db

	if query.QuestionID != nil {
		tx = tx.Where(&model.Option{QuestionID: *query.QuestionID})
	}

	var options []model.Option

	if err := tx.Find(&options).Error; err != nil {
		return nil, err
	}

	return options, nil
}

package gorm

import (
	"context"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
)

func (r *Repository) CreateAnswers(ctx context.Context, answers *[]model.Answer) error {
	if err := gorm.G[[]model.Answer](r.db).Create(ctx, answers); err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetAnswerByID(id uint) (*model.Answer, error) {
	var answer model.Answer

	if err := r.db.First(&answer, id).Error; err != nil {
		return nil, err
	}

	return &answer, nil
}

func (r *Repository) UpdateAnswer(answer *model.Answer) error {
	if err := r.db.Save(answer).Error; err != nil {
		return err
	}

	return nil
}

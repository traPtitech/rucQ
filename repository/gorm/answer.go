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

func (r *Repository) GetAnswersByUserAndQuestionGroup(
	ctx context.Context,
	userID string,
	questionGroupID uint,
) ([]model.Answer, error) {
	answers, err := gorm.G[model.Answer](r.db).
		Where("user_id = ? AND question_id IN (?)", userID,
			r.db.Model(&model.Question{}).
				Select("id").
				Where("question_group_id = ?", questionGroupID),
		).
		Preload("SelectedOptions", nil).
		Find(ctx)

	if err != nil {
		return nil, err
	}

	return answers, nil
}

func (r *Repository) UpdateAnswer(ctx context.Context, answer *model.Answer) error {
	if err := r.db.Save(answer).Error; err != nil {
		return err
	}

	return nil
}

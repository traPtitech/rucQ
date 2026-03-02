package gormrepository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
)

func (r *Repository) CreateQuestionGroup(questionGroup *model.QuestionGroup) error {
	if err := r.db.Create(questionGroup).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetQuestionGroups(
	ctx context.Context,
	campID uint,
) ([]model.QuestionGroup, error) {
	questionGroups, err := gorm.G[model.QuestionGroup](r.db).
		Preload("Questions.Options", nil).
		Where("camp_id = ?", campID).
		Find(ctx)

	if err != nil {
		return nil, err
	}

	return questionGroups, nil
}

func (r *Repository) GetQuestionGroup(ctx context.Context, ID uint) (*model.QuestionGroup, error) {
	questionGroup, err := gorm.G[model.QuestionGroup](r.db).
		Preload("Questions.Options", nil).
		Where("id = ?", ID).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}

		return nil, err
	}

	return &questionGroup, nil
}

func (r *Repository) UpdateQuestionGroup(
	ctx context.Context,
	questionGroupID uint,
	questionGroup model.QuestionGroup,
) error {
	if _, err := gorm.G[model.QuestionGroup](
		r.db,
	).Where("id = ?", questionGroupID).
		Select("name", "description", "due").
		Updates(ctx, questionGroup); err != nil {
		return err
	}

	return nil
}

func (r *Repository) DeleteQuestionGroup(ID uint) error {
	if err := r.db.
		Delete(&model.QuestionGroup{}, ID).
		Error; err != nil {
		return err
	}

	return nil
}

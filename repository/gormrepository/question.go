package gormrepository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
)

func (r *Repository) CreateQuestion(question *model.Question) error {
	if err := r.db.Create(question).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetQuestions() ([]model.Question, error) {
	var questions []model.Question

	if err := r.db.Preload("Options").Find(&questions).Error; err != nil {
		return nil, err
	}

	return questions, nil
}

func (r *Repository) GetQuestionByID(id uint) (*model.Question, error) {
	var question model.Question

	if err := r.db.Preload("Options").First(&question, id).Error; err != nil {
		return nil, err
	}

	return &question, nil
}

func (r *Repository) DeleteQuestionByID(id uint) error {
	if err := r.db.Delete(&model.Question{}, id).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpdateQuestion(
	ctx context.Context,
	questionID uint,
	question *model.Question,
) error {
	question.ID = questionID

	if err := r.db.WithContext(ctx).
		Session(&gorm.Session{FullSaveAssociations: true}).
		Select(
			"type",
			"title",
			"description",
			"is_public",
			"is_open",
			"is_required",
			"Options",
		).
		Updates(question).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetQuestionCampID(ctx context.Context, questionID uint) (uint, error) {
	type campResult struct {
		CampID uint `gorm:"column:camp_id"`
	}

	var result campResult

	err := r.db.WithContext(ctx).
		Table("questions").
		Select("question_groups.camp_id").
		Joins("JOIN question_groups ON question_groups.id = questions.question_group_id").
		Where("questions.id = ?", questionID).
		Where("questions.deleted_at IS NULL AND question_groups.deleted_at IS NULL").
		Take(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, model.ErrNotFound
		}

		return 0, err
	}

	return result.CampID, nil
}

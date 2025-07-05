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

func (r *Repository) GetAnswerByID(ctx context.Context, id uint) (*model.Answer, error) {
	answer, err := gorm.G[model.Answer](r.db).
		Preload("SelectedOptions", nil).
		Where("id = ?", id).
		First(ctx)

	if err != nil {
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

func (r *Repository) UpdateAnswer(ctx context.Context, answerID uint, answer *model.Answer) error {
	answer.ID = answerID

	if _, err := gorm.G[*model.Answer](r.db).Omit("SelectedOptions").Where("id = ?", answerID).Updates(ctx, answer); err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).Model(answer).Association("SelectedOptions").Replace(answer.SelectedOptions); err != nil {
		return err
	}

	// 更新後のデータを取得してanswerに反映
	updatedAnswer, err := r.GetAnswerByID(ctx, answerID)
	if err != nil {
		return err
	}

	*answer = *updatedAnswer

	return nil
}

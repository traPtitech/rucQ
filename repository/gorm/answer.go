package gorm

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
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

func (r *Repository) GetAnswers(
	ctx context.Context,
	query repository.GetAnswersQuery,
) ([]model.Answer, error) {
	db := r.db.WithContext(ctx)

	// QuestionIDが指定されている場合、質問の存在確認を行う
	if query.QuestionID != nil {
		var question model.Question
		if err := r.db.WithContext(ctx).First(&question, *query.QuestionID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, model.ErrNotFound
			}
			return nil, err
		}

		// 非公開回答を含めない場合は、公開質問かチェック
		if !query.IncludePrivateAnswers && !question.IsPublic {
			return nil, model.ErrForbidden
		}
	}

	if query.UserID != nil {
		db = db.Where("user_id = ?", *query.UserID)
	}

	if query.QuestionGroupID != nil {
		db = db.Where("question_id IN (?)",
			r.db.Model(&model.Question{}).
				Select("id").
				Where("question_group_id = ?", *query.QuestionGroupID),
		)
	}

	if query.QuestionID != nil {
		db = db.Where("question_id = ?", *query.QuestionID)
	}

	// 非公開回答を含めない場合は、公開質問のみにフィルタ
	if !query.IncludePrivateAnswers && query.QuestionID == nil {
		db = db.Where("question_id IN (?)",
			r.db.Model(&model.Question{}).
				Select("id").
				Where("is_public = ?", true),
		)
	}

	var answers []model.Answer
	err := db.Preload("SelectedOptions").Find(&answers).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}

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

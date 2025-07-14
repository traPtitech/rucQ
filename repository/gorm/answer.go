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
	if !query.IncludePrivateAnswers {
		if query.QuestionID == nil {
			return nil, errors.New("QuestionID is required")
		}

		question, err := gorm.G[model.Question](r.db).
			Where("id = ?", query.QuestionID).
			First(ctx)

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, model.ErrNotFound
			}

			return nil, err
		}

		if !question.IsPublic {
			return nil, model.ErrForbidden
		}
	}

	scopes := make([]func(*gorm.Statement), 0, 3)

	if query.UserID != nil {
		scopes = append(scopes, func(s *gorm.Statement) {
			s.Where("user_id = ?", query.UserID)
		})
	}

	if query.QuestionGroupID != nil {
		scopes = append(scopes, func(s *gorm.Statement) {
			s.Where("question_id IN (?)",
				s.DB.Model(&model.Question{}).
					Select("id").
					Where("question_group_id = ?", query.QuestionGroupID),
			)
		})
	}

	if query.QuestionID != nil {
		scopes = append(scopes, func(s *gorm.Statement) {
			s.Where("question_id = ?", query.QuestionID)
		})
	}

	answers, err := gorm.G[model.Answer](r.db).
		Scopes(scopes...).
		Preload("SelectedOptions", nil).
		Find(ctx)

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

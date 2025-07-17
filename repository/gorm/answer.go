package gorm

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
)

func (r *Repository) CreateAnswer(ctx context.Context, answer *model.Answer) error {
	if err := gorm.G[model.Answer](r.db).Create(ctx, answer); err != nil {
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			return model.ErrNotFound
		}

		return err
	}

	// 選択肢を反映するため再取得する
	newAnswer, err := r.GetAnswerByID(ctx, answer.ID)

	if err != nil {
		return err
	}

	*answer = *newAnswer

	return nil
}

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
	// QuestionIDが指定されている場合、質問の存在確認を行う
	if query.QuestionID != nil {
		question, err := gorm.G[model.Question](r.db).
			Where("id = ?", query.QuestionID).
			First(ctx)

		if err != nil {
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

	scopes := make([]func(*gorm.Statement), 0, 3)

	if query.UserID != nil {
		scopes = append(scopes, func(s *gorm.Statement) {
			s.Where("user_id = ?", query.UserID)
		})

	}

	if query.QuestionGroupID != nil {
		questionGroup, err := r.GetQuestionGroup(ctx, *query.QuestionGroupID)

		if err != nil {
			return nil, err
		}

		questionIDs := make([]uint, len(questionGroup.Questions))

		for i, question := range questionGroup.Questions {
			questionIDs[i] = question.ID
		}

		scopes = append(scopes, func(s *gorm.Statement) {
			s.Where("question_id IN (?)",
				questionIDs,
			)
		})

	}

	if query.QuestionID != nil {
		scopes = append(scopes, func(s *gorm.Statement) {
			s.Where("question_id = ?", query.QuestionID)
		})
	}

	// 非公開回答を含めない場合は、公開質問のみにフィルタ
	if !query.IncludePrivateAnswers && query.QuestionID == nil {
		publicQuestions, err := gorm.G[model.Question](r.db).
			Select("id").
			Where("is_public = ?", true).
			Find(ctx)

		if err != nil {
			return nil, err
		}

		publicQuestionIDs := make([]uint, len(publicQuestions))

		for i, question := range publicQuestions {
			publicQuestionIDs[i] = question.ID
		}

		scopes = append(scopes, func(s *gorm.Statement) {
			s.Where("question_id IN (?)",
				publicQuestionIDs,
			)
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

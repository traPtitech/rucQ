package gormrepository

import (
	"context"
	"errors"
	"slices"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
)

func (r *Repository) CreateAnswer(ctx context.Context, answer *model.Answer) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := &Repository{db: tx}
		if err := repo.ValidateAnswer(ctx, []model.Answer{*answer}); err != nil {
			return err
		}
		if err := gorm.G[model.Answer](tx).Create(ctx, answer); err != nil {
			if errors.Is(err, gorm.ErrForeignKeyViolated) {
				return model.ErrNotFound
			}
			return err
		}
		// 選択肢を反映するため再取得する
		newAnswer, err := repo.GetAnswerByID(ctx, answer.ID)
		if err != nil {
			return err
		}

		*answer = *newAnswer

		return nil
	})
}

func (r *Repository) CreateAnswers(ctx context.Context, answers *[]model.Answer) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := &Repository{db: tx}
		if err := repo.ValidateAnswer(ctx, *answers); err != nil {
			return err
		}
		return gorm.G[[]model.Answer](tx).Create(ctx, answers)
	})
}

func (r *Repository) ValidateAnswer(ctx context.Context, answers []model.Answer) error {
	if len(answers) == 0 {
		return nil
	}
	questionIDs := make([]uint, 0, len(answers))
	seen := make(map[uint]struct{}, len(answers))
	for _, answer := range answers {
		if _, ok := seen[answer.QuestionID]; !ok {
			seen[answer.QuestionID] = struct{}{}
			questionIDs = append(questionIDs, answer.QuestionID)
		}
	}

	slices.Sort(questionIDs)
	var questions []model.Question
	if err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Select("id", "is_open").Where("id IN ?", questionIDs).
		Order("id").Find(&questions).Error; err != nil {
		return err
	}
	if len(questions) != len(questionIDs) {
		return model.ErrNotFound
	}
	for _, question := range questions {
		if !question.IsOpen {
			return repository.ErrQuestionClosed
		}
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

	const maxScopes = 4

	scopes := make([]func(*gorm.Statement), 0, maxScopes)

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

	// 参加者以外の回答を含めない場合
	if !query.IncludeNonParticipants {
		scopes = append(scopes, func(s *gorm.Statement) {
			subQuery := r.db.Table("camp_participants").
				Select("1").
				Joins("JOIN question_groups ON question_groups.camp_id = camp_participants.camp_id").
				Joins("JOIN questions ON questions.question_group_id = question_groups.id").
				Where("questions.id = answers.question_id AND camp_participants.user_id = answers.user_id")

			s.Where("EXISTS (?)", subQuery)
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
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := &Repository{db: tx}
		existing, err := repo.GetAnswerByID(ctx, answerID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return model.ErrNotFound
			}
			return err
		}
		// Validate the stored question, not the question ID supplied by the client.
		if err := repo.ValidateAnswer(ctx, []model.Answer{*existing}); err != nil {
			return err
		}
		answer.ID = answerID
		if _, err := gorm.G[*model.Answer](tx).
			Omit("SelectedOptions", "QuestionID", "UserID").
			Where("id = ?", answerID).Updates(ctx, answer); err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Model(answer).
			Association("SelectedOptions").Replace(answer.SelectedOptions); err != nil {
			return err
		}
		updatedAnswer, err := repo.GetAnswerByID(ctx, answerID)
		if err != nil {
			return err
		}
		*answer = *updatedAnswer
		return nil
	})
}

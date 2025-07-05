package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestCreateAnswers(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)
		questionGroup := mustCreateQuestionGroup(t, r, camp.ID)
		freeTextQuestion := mustCreateQuestion(t, r, questionGroup.ID, model.FreeTextQuestion)
		freeTextContent := random.AlphaNumericString(t, 20)
		freeNumberQuestion := mustCreateQuestion(t, r, questionGroup.ID, model.FreeNumberQuestion)
		freeNumberContent := random.Float64(t)
		singleChoiceQuestion := mustCreateQuestion(
			t,
			r,
			questionGroup.ID,
			model.SingleChoiceQuestion,
		)
		multipleChoiceQuestion := mustCreateQuestion(
			t,
			r,
			questionGroup.ID,
			model.MultipleChoiceQuestion,
		)
		answers := []model.Answer{
			{
				QuestionID:      freeTextQuestion.ID,
				UserID:          user.ID,
				Type:            model.FreeTextQuestion,
				FreeTextContent: &freeTextContent,
			},
			{
				QuestionID:        freeNumberQuestion.ID,
				UserID:            user.ID,
				Type:              model.FreeNumberQuestion,
				FreeNumberContent: &freeNumberContent,
			},
			{
				QuestionID: singleChoiceQuestion.ID,
				UserID:     user.ID,
				Type:       model.SingleChoiceQuestion,
				SelectedOptions: []model.Option{
					singleChoiceQuestion.Options[0],
				},
			},
			{
				QuestionID:      multipleChoiceQuestion.ID,
				UserID:          user.ID,
				Type:            model.MultipleChoiceQuestion,
				SelectedOptions: multipleChoiceQuestion.Options,
			},
		}

		err := r.CreateAnswers(t.Context(), &answers)

		assert.NoError(t, err)
	})
}

func TestGetAnswersByUserAndQuestionGroup(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)
		questionGroup := mustCreateQuestionGroup(t, r, camp.ID)

		// 別のquestion groupも作成して、確実にフィルタリングされるかテスト
		anotherQuestionGroup := mustCreateQuestionGroup(t, r, camp.ID)

		freeTextQuestion := mustCreateQuestion(t, r, questionGroup.ID, model.FreeTextQuestion)
		freeNumberQuestion := mustCreateQuestion(t, r, questionGroup.ID, model.FreeNumberQuestion)

		// 別のquestion groupの質問も作成
		anotherQuestion := mustCreateQuestion(t, r, anotherQuestionGroup.ID, model.FreeTextQuestion)

		freeTextContent := random.AlphaNumericString(t, 20)
		freeNumberContent := random.Float64(t)
		anotherContent := random.AlphaNumericString(t, 20)

		answers := []model.Answer{
			{
				QuestionID:      freeTextQuestion.ID,
				UserID:          user.ID,
				Type:            model.FreeTextQuestion,
				FreeTextContent: &freeTextContent,
			},
			{
				QuestionID:        freeNumberQuestion.ID,
				UserID:            user.ID,
				Type:              model.FreeNumberQuestion,
				FreeNumberContent: &freeNumberContent,
			},
			{
				QuestionID:      anotherQuestion.ID,
				UserID:          user.ID,
				Type:            model.FreeTextQuestion,
				FreeTextContent: &anotherContent,
			},
		}

		err := r.CreateAnswers(t.Context(), &answers)
		assert.NoError(t, err)

		// 特定のquestion groupの回答のみ取得
		result, err := r.GetAnswersByUserAndQuestionGroup(t.Context(), user.ID, questionGroup.ID)

		assert.NoError(t, err)
		assert.Len(t, result, 2) // 2つの回答のみ取得されるはず

		// 取得した回答が正しいquestion groupのものか確認
		for _, answer := range result {
			assert.Equal(t, user.ID, answer.UserID)
			assert.True(
				t,
				answer.QuestionID == freeTextQuestion.ID ||
					answer.QuestionID == freeNumberQuestion.ID,
			)
		}
	})

	t.Run("No Answers", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)
		questionGroup := mustCreateQuestionGroup(t, r, camp.ID)

		result, err := r.GetAnswersByUserAndQuestionGroup(t.Context(), user.ID, questionGroup.ID)

		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Different User", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user1 := mustCreateUser(t, r)
		user2 := mustCreateUser(t, r)
		questionGroup := mustCreateQuestionGroup(t, r, camp.ID)

		freeTextQuestion := mustCreateQuestion(t, r, questionGroup.ID, model.FreeTextQuestion)
		freeTextContent := random.AlphaNumericString(t, 20)

		answers := []model.Answer{
			{
				QuestionID:      freeTextQuestion.ID,
				UserID:          user1.ID,
				Type:            model.FreeTextQuestion,
				FreeTextContent: &freeTextContent,
			},
		}

		err := r.CreateAnswers(t.Context(), &answers)
		assert.NoError(t, err)

		// 別のユーザーで検索
		result, err := r.GetAnswersByUserAndQuestionGroup(t.Context(), user2.ID, questionGroup.ID)

		assert.NoError(t, err)
		assert.Empty(t, result) // 別のユーザーの回答は取得されない
	})
}

func TestUpdateAnswer(t *testing.T) {
	t.Parallel()

	t.Run("Success with SingleChoiceQuestion", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)
		questionGroup := mustCreateQuestionGroup(t, r, camp.ID)
		singleChoiceQuestion := mustCreateQuestion(
			t,
			r,
			questionGroup.ID,
			model.SingleChoiceQuestion,
		)

		answers := []model.Answer{
			{
				QuestionID: singleChoiceQuestion.ID,
				UserID:     user.ID,
				Type:       model.SingleChoiceQuestion,
				SelectedOptions: []model.Option{
					singleChoiceQuestion.Options[0],
				},
			},
		}

		err := r.CreateAnswers(t.Context(), &answers)
		require.NoError(t, err)

		// CreateAnswers後に作成されたanswerのIDを取得
		createdAnswer := answers[0]
		require.NotZero(t, createdAnswer.ID)

		// 選択肢を変更してアップデート
		createdAnswer.SelectedOptions = []model.Option{singleChoiceQuestion.Options[1]}

		err = r.UpdateAnswer(t.Context(), createdAnswer.ID, &createdAnswer)
		assert.NoError(t, err)

		retrievedAnswer, err := r.GetAnswerByID(t.Context(), createdAnswer.ID)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(retrievedAnswer.SelectedOptions))
		assert.Equal(t, createdAnswer.SelectedOptions[0].ID, retrievedAnswer.SelectedOptions[0].ID)
	})
}

package gormrepository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

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
		createdAnswer.SelectedOptions = []model.Option{
			{
				Model: gorm.Model{
					ID: singleChoiceQuestion.Options[1].ID,
				},
				Content: "", // IDのみを指定して更新
			},
		}

		err = r.UpdateAnswer(t.Context(), createdAnswer.ID, &createdAnswer)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(createdAnswer.SelectedOptions))
		assert.Equal(t, singleChoiceQuestion.Options[1].ID, createdAnswer.SelectedOptions[0].ID)
		assert.Equal(
			t,
			singleChoiceQuestion.Options[1].Content,
			createdAnswer.SelectedOptions[0].Content,
		)
	})
}

func TestGetAnswersByQuestionID(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user1 := mustCreateUser(t, r)
		user2 := mustCreateUser(t, r)
		questionGroup := mustCreateQuestionGroup(t, r, camp.ID)
		question := mustCreateQuestion(t, r, questionGroup.ID, model.FreeTextQuestion)
		freeTextContent1 := random.AlphaNumericString(t, 20)
		freeTextContent2 := random.AlphaNumericString(t, 20)

		// 同じ質問に対する複数のAnswerを作成
		answers := []model.Answer{
			{
				QuestionID:      question.ID,
				UserID:          user1.ID,
				Type:            model.FreeTextQuestion,
				FreeTextContent: &freeTextContent1,
			},
			{
				QuestionID:      question.ID,
				UserID:          user2.ID,
				Type:            model.FreeTextQuestion,
				FreeTextContent: &freeTextContent2,
			},
		}

		err := r.CreateAnswers(t.Context(), &answers)
		require.NoError(t, err)

		// QuestionIDでAnswerを取得
		retrievedAnswers, err := r.GetAnswersByQuestionID(t.Context(), question.ID)
		assert.NoError(t, err)
		assert.Len(t, retrievedAnswers, 2)

		// 結果を検証
		answerMap := make(map[string]model.Answer)
		for _, answer := range retrievedAnswers {
			answerMap[answer.UserID] = answer
		}

		assert.Contains(t, answerMap, user1.ID)
		assert.Contains(t, answerMap, user2.ID)
		assert.Equal(t, question.ID, answerMap[user1.ID].QuestionID)
		assert.Equal(t, question.ID, answerMap[user2.ID].QuestionID)
		assert.Equal(t, freeTextContent1, *answerMap[user1.ID].FreeTextContent)
		assert.Equal(t, freeTextContent2, *answerMap[user2.ID].FreeTextContent)
	})

	t.Run("No Answers", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		questionGroup := mustCreateQuestionGroup(t, r, camp.ID)
		question := mustCreateQuestion(t, r, questionGroup.ID, model.FreeTextQuestion)

		// Answerが存在しない質問に対するクエリ
		answers, err := r.GetAnswersByQuestionID(t.Context(), question.ID)
		assert.NoError(t, err)
		assert.Empty(t, answers)
	})

	t.Run("With SelectedOptions", func(t *testing.T) {
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

		// QuestionIDでAnswerを取得（SelectedOptionsも含む）
		retrievedAnswers, err := r.GetAnswersByQuestionID(t.Context(), singleChoiceQuestion.ID)
		assert.NoError(t, err)
		assert.Len(t, retrievedAnswers, 1)
		assert.Len(t, retrievedAnswers[0].SelectedOptions, 1)
		assert.Equal(
			t,
			singleChoiceQuestion.Options[0].ID,
			retrievedAnswers[0].SelectedOptions[0].ID,
		)
		assert.Equal(
			t,
			singleChoiceQuestion.Options[0].Content,
			retrievedAnswers[0].SelectedOptions[0].Content,
		)
	})
}

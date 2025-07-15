package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
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
		freeTextQuestion := mustCreateQuestion(t, r, questionGroup.ID, model.FreeTextQuestion, nil)
		freeTextContent := random.AlphaNumericString(t, 20)
		freeNumberQuestion := mustCreateQuestion(
			t,
			r,
			questionGroup.ID,
			model.FreeNumberQuestion,
			nil,
		)
		freeNumberContent := random.Float64(t)
		singleChoiceQuestion := mustCreateQuestion(
			t,
			r,
			questionGroup.ID,
			model.SingleChoiceQuestion,
			nil,
		)
		multipleChoiceQuestion := mustCreateQuestion(
			t,
			r,
			questionGroup.ID,
			model.MultipleChoiceQuestion,
			nil,
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

		freeTextQuestion := mustCreateQuestion(t, r, questionGroup.ID, model.FreeTextQuestion, nil)
		freeNumberQuestion := mustCreateQuestion(
			t,
			r,
			questionGroup.ID,
			model.FreeNumberQuestion,
			nil,
		)

		// 別のquestion groupの質問も作成
		anotherQuestion := mustCreateQuestion(
			t,
			r,
			anotherQuestionGroup.ID,
			model.FreeTextQuestion,
			nil,
		)

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
		query := repository.GetAnswersQuery{
			UserID:          &user.ID,
			QuestionGroupID: &questionGroup.ID,
		}
		result, err := r.GetAnswers(t.Context(), query)

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

		query := repository.GetAnswersQuery{
			UserID:          &user.ID,
			QuestionGroupID: &questionGroup.ID,
		}
		result, err := r.GetAnswers(t.Context(), query)

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

		freeTextQuestion := mustCreateQuestion(t, r, questionGroup.ID, model.FreeTextQuestion, nil)
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
		query := repository.GetAnswersQuery{
			UserID:          &user2.ID,
			QuestionGroupID: &questionGroup.ID,
		}
		result, err := r.GetAnswers(t.Context(), query)

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
			nil,
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
		question := mustCreateQuestion(t, r, questionGroup.ID, model.FreeTextQuestion, nil)
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
		query := repository.GetAnswersQuery{
			QuestionID:            &question.ID,
			IncludePrivateAnswers: true,
		}
		retrievedAnswers, err := r.GetAnswers(t.Context(), query)
		assert.NoError(t, err)
		assert.Len(t, retrievedAnswers, 2)

		// 結果を検証
		answerMap := make(map[string]model.Answer)
		for _, answer := range retrievedAnswers {
			answerMap[answer.UserID] = answer
		}

		if assert.Contains(t, answerMap, user1.ID) {
			assert.Equal(t, question.ID, answerMap[user1.ID].QuestionID)
			assert.Equal(t, freeTextContent1, *answerMap[user1.ID].FreeTextContent)
		}

		if assert.Contains(t, answerMap, user2.ID) {
			assert.Equal(t, question.ID, answerMap[user2.ID].QuestionID)
			assert.Equal(t, freeTextContent2, *answerMap[user2.ID].FreeTextContent)
		}
	})

	t.Run("No Answers", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		questionGroup := mustCreateQuestionGroup(t, r, camp.ID)
		question := mustCreateQuestion(t, r, questionGroup.ID, model.FreeTextQuestion, nil)

		// Answerが存在しない質問に対するクエリ
		query := repository.GetAnswersQuery{
			QuestionID:            &question.ID,
			IncludePrivateAnswers: true,
		}
		answers, err := r.GetAnswers(t.Context(), query)
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
			nil,
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
		query := repository.GetAnswersQuery{
			QuestionID: &singleChoiceQuestion.ID,
		}
		retrievedAnswers, err := r.GetAnswers(t.Context(), query)
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

func TestGetPublicAnswersByQuestionID(t *testing.T) {
	t.Parallel()

	t.Run("Success - Public Question", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user1 := mustCreateUser(t, r)
		user2 := mustCreateUser(t, r)
		questionGroup := mustCreateQuestionGroup(t, r, camp.ID)

		// Public質問を作成
		isPublic := true
		publicQuestion := mustCreateQuestion(
			t,
			r,
			questionGroup.ID,
			model.FreeTextQuestion,
			&isPublic,
		)
		freeTextContent1 := random.AlphaNumericString(t, 20)
		freeTextContent2 := random.AlphaNumericString(t, 20)

		// 同じ質問に対する複数のAnswerを作成
		answers := []model.Answer{
			{
				QuestionID:      publicQuestion.ID,
				UserID:          user1.ID,
				Type:            model.FreeTextQuestion,
				FreeTextContent: &freeTextContent1,
			},
			{
				QuestionID:      publicQuestion.ID,
				UserID:          user2.ID,
				Type:            model.FreeTextQuestion,
				FreeTextContent: &freeTextContent2,
			},
		}

		err := r.CreateAnswers(t.Context(), &answers)
		require.NoError(t, err)

		// Public質問の回答を取得
		query := repository.GetAnswersQuery{
			QuestionID:            &publicQuestion.ID,
			IncludePrivateAnswers: false,
		}
		retrievedAnswers, err := r.GetAnswers(t.Context(), query)
		assert.NoError(t, err)
		assert.Len(t, retrievedAnswers, 2)

		// 結果を検証
		answerMap := make(map[string]model.Answer)
		for _, answer := range retrievedAnswers {
			answerMap[answer.UserID] = answer
		}

		assert.Contains(t, answerMap, user1.ID)
		assert.Contains(t, answerMap, user2.ID)
		assert.Equal(t, publicQuestion.ID, answerMap[user1.ID].QuestionID)
		assert.Equal(t, publicQuestion.ID, answerMap[user2.ID].QuestionID)
		assert.Equal(t, freeTextContent1, *answerMap[user1.ID].FreeTextContent)
		assert.Equal(t, freeTextContent2, *answerMap[user2.ID].FreeTextContent)
	})

	t.Run("Private Question - Returns Empty", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)
		questionGroup := mustCreateQuestionGroup(t, r, camp.ID)

		// Private質問を作成
		isPublic := false
		privateQuestion := mustCreateQuestion(
			t,
			r,
			questionGroup.ID,
			model.FreeTextQuestion,
			&isPublic,
		)
		freeTextContent := random.AlphaNumericString(t, 20)

		// Private質問に回答を作成
		answers := []model.Answer{
			{
				QuestionID:      privateQuestion.ID,
				UserID:          user.ID,
				Type:            model.FreeTextQuestion,
				FreeTextContent: &freeTextContent,
			},
		}

		err := r.CreateAnswers(t.Context(), &answers)
		require.NoError(t, err)

		// Private質問の回答を取得
		query := repository.GetAnswersQuery{
			QuestionID:            &privateQuestion.ID,
			IncludePrivateAnswers: false,
		}
		retrievedAnswers, err := r.GetAnswers(t.Context(), query)

		if assert.Error(t, err) {
			assert.Equal(t, model.ErrForbidden, err)
			assert.Empty(t, retrievedAnswers)
		}
	})

	t.Run("Non-existent Question", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		// 存在しない質問IDで回答を取得
		nonExistentQuestionID := uint(99999)
		query := repository.GetAnswersQuery{
			QuestionID:            &nonExistentQuestionID,
			IncludePrivateAnswers: false,
		}
		answers, err := r.GetAnswers(t.Context(), query)

		if assert.Error(t, err) {
			assert.Equal(t, model.ErrNotFound, err)
			assert.Empty(t, answers)
		}
	})

	t.Run("Public Question with SelectedOptions", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)
		questionGroup := mustCreateQuestionGroup(t, r, camp.ID)

		// Public選択肢質問を作成
		isPublic := true
		publicSingleChoiceQuestion := mustCreateQuestion(
			t,
			r,
			questionGroup.ID,
			model.SingleChoiceQuestion,
			&isPublic,
		)

		answers := []model.Answer{
			{
				QuestionID: publicSingleChoiceQuestion.ID,
				UserID:     user.ID,
				Type:       model.SingleChoiceQuestion,
				SelectedOptions: []model.Option{
					publicSingleChoiceQuestion.Options[0],
				},
			},
		}

		err := r.CreateAnswers(t.Context(), &answers)
		require.NoError(t, err)

		// Public質問の回答を取得（SelectedOptionsも含む）
		query := repository.GetAnswersQuery{
			QuestionID:            &publicSingleChoiceQuestion.ID,
			IncludePrivateAnswers: false,
		}
		retrievedAnswers, err := r.GetAnswers(t.Context(), query)
		assert.NoError(t, err)
		assert.Len(t, retrievedAnswers, 1)
		assert.Len(t, retrievedAnswers[0].SelectedOptions, 1)
		assert.Equal(
			t,
			publicSingleChoiceQuestion.Options[0].ID,
			retrievedAnswers[0].SelectedOptions[0].ID,
		)
		assert.Equal(
			t,
			publicSingleChoiceQuestion.Options[0].Content,
			retrievedAnswers[0].SelectedOptions[0].Content,
		)
	})
}

func TestGetAnswersByQuestionGroup(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user1 := mustCreateUser(t, r)
		user2 := mustCreateUser(t, r)
		questionGroup := mustCreateQuestionGroup(t, r, camp.ID)
		question1 := mustCreateQuestion(t, r, questionGroup.ID, model.FreeTextQuestion, nil)
		question2 := mustCreateQuestion(t, r, questionGroup.ID, model.FreeTextQuestion, nil)

		// Create answers from different users for the same question group
		content1 := random.AlphaNumericString(t, 20)
		content2 := random.AlphaNumericString(t, 20)
		content3 := random.AlphaNumericString(t, 20)

		answers := []model.Answer{
			{
				QuestionID:      question1.ID,
				UserID:          user1.ID,
				Type:            model.FreeTextQuestion,
				FreeTextContent: &content1,
			},
			{
				QuestionID:      question2.ID,
				UserID:          user1.ID,
				Type:            model.FreeTextQuestion,
				FreeTextContent: &content2,
			},
			{
				QuestionID:      question1.ID,
				UserID:          user2.ID,
				Type:            model.FreeTextQuestion,
				FreeTextContent: &content3,
			},
		}

		require.NoError(t, r.CreateAnswers(t.Context(), &answers))

		// Get all answers for the question group
		query := repository.GetAnswersQuery{
			QuestionGroupID: &questionGroup.ID,
		}
		retrievedAnswers, err := r.GetAnswers(t.Context(), query)
		require.NoError(t, err)
		require.Len(t, retrievedAnswers, 3)

		// Verify that answers from both users are returned
		userIDs := make(map[string]bool)
		questionIDs := make(map[uint]bool)
		for _, answer := range retrievedAnswers {
			userIDs[answer.UserID] = true
			questionIDs[answer.QuestionID] = true
		}

		assert.True(t, userIDs[user1.ID])
		assert.True(t, userIDs[user2.ID])
		assert.True(t, questionIDs[question1.ID])
		assert.True(t, questionIDs[question2.ID])
	})

	t.Run("Empty result when no answers exist", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		questionGroup := mustCreateQuestionGroup(t, r, camp.ID)

		query := repository.GetAnswersQuery{
			QuestionGroupID: &questionGroup.ID,
		}
		retrievedAnswers, err := r.GetAnswers(t.Context(), query)
		require.NoError(t, err)
		assert.Empty(t, retrievedAnswers)
	})

	t.Run("Only returns answers for specified question group", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)
		questionGroup1 := mustCreateQuestionGroup(t, r, camp.ID)
		questionGroup2 := mustCreateQuestionGroup(t, r, camp.ID)
		question1 := mustCreateQuestion(t, r, questionGroup1.ID, model.FreeTextQuestion, nil)
		question2 := mustCreateQuestion(t, r, questionGroup2.ID, model.FreeTextQuestion, nil)

		content1 := random.AlphaNumericString(t, 20)
		content2 := random.AlphaNumericString(t, 20)

		answers := []model.Answer{
			{
				QuestionID:      question1.ID,
				UserID:          user.ID,
				Type:            model.FreeTextQuestion,
				FreeTextContent: &content1,
			},
			{
				QuestionID:      question2.ID,
				UserID:          user.ID,
				Type:            model.FreeTextQuestion,
				FreeTextContent: &content2,
			},
		}

		require.NoError(t, r.CreateAnswers(t.Context(), &answers))

		// Get answers for only question group 1
		query := repository.GetAnswersQuery{
			QuestionGroupID: &questionGroup1.ID,
		}
		retrievedAnswers, err := r.GetAnswers(t.Context(), query)
		require.NoError(t, err)
		require.Len(t, retrievedAnswers, 1)
		assert.Equal(t, question1.ID, retrievedAnswers[0].QuestionID)
		assert.Equal(t, content1, *retrievedAnswers[0].FreeTextContent)
	})
}

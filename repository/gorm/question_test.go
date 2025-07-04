package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestCreateQuestion(t *testing.T) {
	t.Parallel()

	t.Run("Success (Single Choice)", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		questionGroup := mustCreateQuestionGroup(t, r, camp.ID)
		title := random.AlphaNumericString(t, 10)
		description := random.PtrOrNil(t, random.AlphaNumericString(t, 20))
		isPublic := random.Bool(t)
		isOpen := random.Bool(t)
		optionContent := random.AlphaNumericString(t, 10)
		question := model.Question{
			Type:            model.SingleChoiceQuestion,
			QuestionGroupID: questionGroup.ID,
			Title:           title,
			Description:     description,
			IsPublic:        isPublic,
			IsOpen:          isOpen,
			Options: []model.Option{
				{
					Content: optionContent,
				},
			},
		}
		err := r.CreateQuestion(&question)

		assert.NoError(t, err)
		assert.NotZero(t, question.ID)
		assert.Equal(t, questionGroup.ID, question.QuestionGroupID)
		assert.Equal(t, model.SingleChoiceQuestion, question.Type)
		assert.Equal(t, title, question.Title)
		assert.Equal(t, description, question.Description)
		assert.Equal(t, isPublic, question.IsPublic)
		assert.Equal(t, isOpen, question.IsOpen)
		assert.Len(t, question.Options, 1)
		assert.NotZero(t, question.Options[0].ID)
		assert.Equal(t, optionContent, question.Options[0].Content)
	})
}

func TestUpdateQuestion(t *testing.T) {
	t.Parallel()

	t.Run("Success (Single Choice)", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		questionGroup := mustCreateQuestionGroup(t, r, camp.ID)
		question := mustCreateQuestion(t, r, questionGroup.ID, model.SingleChoiceQuestion)

		newTitle := random.AlphaNumericString(t, 15)
		newDescription := random.PtrOrNil(t, random.AlphaNumericString(t, 25))
		newIsPublic := random.Bool(t)
		newIsOpen := random.Bool(t)
		newOptionContent := random.AlphaNumericString(t, 15)
		newQuestion := model.Question{
			Type:            model.SingleChoiceQuestion,
			QuestionGroupID: questionGroup.ID,
			Title:           newTitle,
			Description:     newDescription,
			IsPublic:        newIsPublic,
			IsOpen:          newIsOpen,
			Options: []model.Option{
				{
					Model: gorm.Model{
						ID: question.Options[0].ID, // 既存のOptionを更新する
					},
					Content: newOptionContent,
				},
			},
		}

		err := r.UpdateQuestion(t.Context(), question.ID, newQuestion)

		assert.NoError(t, err)

		retrievedQuestion, err := r.GetQuestionByID(question.ID)

		require.NoError(t, err)

		assert.Equal(t, question.ID, retrievedQuestion.ID)
		assert.Equal(t, questionGroup.ID, retrievedQuestion.QuestionGroupID)
		assert.Equal(t, model.SingleChoiceQuestion, retrievedQuestion.Type)
		assert.Equal(t, newTitle, retrievedQuestion.Title)
		assert.Equal(t, newDescription, retrievedQuestion.Description)
		assert.Equal(t, newIsPublic, retrievedQuestion.IsPublic)
		assert.Equal(t, newIsOpen, retrievedQuestion.IsOpen)
		assert.Len(t, retrievedQuestion.Options, 1)
		assert.Equal(t, question.Options[0].ID, retrievedQuestion.Options[0].ID)
		assert.Equal(t, newOptionContent, retrievedQuestion.Options[0].Content)
	})
}

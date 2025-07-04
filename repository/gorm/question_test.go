package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

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

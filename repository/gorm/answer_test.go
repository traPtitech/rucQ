package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
		singleChoiceQuestion := mustCreateQuestion(t, r, questionGroup.ID, model.SingleChoiceQuestion)
		multipleChoiceQuestion := mustCreateQuestion(t, r, questionGroup.ID, model.MultipleChoiceQuestion)
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

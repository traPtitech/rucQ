package gorm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestCreateQuestionGroup(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		name := random.AlphaNumericString(t, 20)
		description := random.PtrOrNil(t, random.AlphaNumericString(t, 100))
		due := random.Time(t)
		questions := []model.Question{
			{
				Type:        model.FreeTextQuestion,
				Title:       random.AlphaNumericString(t, 20),
				Description: random.PtrOrNil(t, random.AlphaNumericString(t, 100)),
				IsPublic:    random.Bool(t),
				IsOpen:      random.Bool(t),
			},
			{
				Type:        model.FreeNumberQuestion,
				Title:       random.AlphaNumericString(t, 20),
				Description: random.PtrOrNil(t, random.AlphaNumericString(t, 100)),
				IsPublic:    random.Bool(t),
				IsOpen:      random.Bool(t),
			},
			{
				Type:        model.SingleChoiceQuestion,
				Title:       random.AlphaNumericString(t, 20),
				Description: random.PtrOrNil(t, random.AlphaNumericString(t, 100)),
				IsPublic:    random.Bool(t),
				IsOpen:      random.Bool(t),
				Options: []model.Option{
					{
						Content: random.AlphaNumericString(t, 20),
					},
				},
			},
			{
				Type:        model.MultipleChoiceQuestion,
				Title:       random.AlphaNumericString(t, 20),
				Description: random.PtrOrNil(t, random.AlphaNumericString(t, 100)),
				IsPublic:    random.Bool(t),
				IsOpen:      random.Bool(t),
				Options: []model.Option{
					{
						Content: random.AlphaNumericString(t, 20),
					},
					{
						Content: random.AlphaNumericString(t, 20),
					},
				},
			},
		}
		camp := mustCreateCamp(t, r)
		questionGroup := model.QuestionGroup{
			Name:        name,
			Description: description,
			Due:         due,
			Questions:   questions,
			CampID:      camp.ID,
		}

		err := r.CreateQuestionGroup(&questionGroup)

		assert.NoError(t, err)

		createdQuestionGroup, err := r.GetQuestionGroup(questionGroup.ID)

		assert.NoError(t, err)
		assert.NotZero(t, createdQuestionGroup.ID)
		assert.Equal(t, name, createdQuestionGroup.Name)
		assert.Equal(t, description, createdQuestionGroup.Description)
		assert.WithinDuration(t, due, createdQuestionGroup.Due, time.Second)
		assert.Equal(t, camp.ID, createdQuestionGroup.CampID)
		assert.Len(t, createdQuestionGroup.Questions, len(questions))

		// 各Questionの検証
		for i, question := range questions {
			createdQuestion := createdQuestionGroup.Questions[i]

			assert.NotZero(t, createdQuestion.ID)
			assert.Equal(t, question.Type, createdQuestion.Type)
			assert.Equal(t, question.Title, createdQuestion.Title)
			assert.Equal(t, question.Description, createdQuestion.Description)
			assert.Equal(t, question.IsPublic, createdQuestion.IsPublic)
			assert.Equal(t, question.IsOpen, createdQuestion.IsOpen)
			assert.Equal(t, createdQuestionGroup.ID, createdQuestion.QuestionGroupID)

			// Optionsの検証
			assert.Len(t, createdQuestion.Options, len(question.Options))

			for j, originalOption := range question.Options {
				createdOption := createdQuestion.Options[j]

				assert.NotZero(t, createdOption.ID)
				assert.Equal(t, originalOption.Content, createdOption.Content)
				assert.Equal(t, createdQuestion.ID, createdOption.QuestionID)
			}
		}
	})
}

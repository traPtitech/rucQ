package gorm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestGetQuestionGroups(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		questionGroup1 := mustCreateQuestionGroup(t, r, camp.ID)
		question := mustCreateQuestion(t, r, questionGroup1.ID, model.SingleChoiceQuestion)
		questionGroup2 := mustCreateQuestionGroup(t, r, camp.ID)

		questionGroups, err := r.GetQuestionGroups(t.Context(), camp.ID)

		assert.NoError(t, err)
		assert.Len(t, questionGroups, 2)
		assert.NotZero(t, questionGroup1.ID)
		assert.Equal(t, questionGroup1.Name, questionGroups[0].Name)
		assert.Equal(t, questionGroup1.Description, questionGroups[0].Description)
		assert.Equal(t, questionGroup1.Due, questionGroups[0].Due)
		assert.Equal(t, camp.ID, questionGroups[0].CampID)
		assert.Len(t, questionGroups[0].Questions, 1)
		assert.NotZero(t, questionGroups[0].Questions[0].ID)
		assert.Equal(t, question.Type, questionGroups[0].Questions[0].Type)
		assert.Equal(t, question.Title, questionGroups[0].Questions[0].Title)
		assert.Equal(t, question.Description, questionGroups[0].Questions[0].Description)
		assert.Equal(t, question.IsPublic, questionGroups[0].Questions[0].IsPublic)
		assert.Equal(t, question.IsOpen, questionGroups[0].Questions[0].IsOpen)
		assert.Equal(t, questionGroup1.ID, questionGroups[0].Questions[0].QuestionGroupID)
		assert.NotZero(t, len(questionGroups[0].Questions[0].Options))

		for _, option := range questionGroups[0].Questions[0].Options {
			assert.NotZero(t, option.ID)
			assert.NotEmpty(t, option.Content)
			assert.Equal(t, questionGroups[0].Questions[0].ID, option.QuestionID)
		}

		assert.NotZero(t, questionGroup2.ID)
		assert.Equal(t, questionGroup2.Name, questionGroups[1].Name)
		assert.Equal(t, questionGroup2.Description, questionGroups[1].Description)
		assert.Equal(t, questionGroup2.Due, questionGroups[1].Due)
		assert.Equal(t, camp.ID, questionGroups[1].CampID)
		assert.Len(t, questionGroups[1].Questions, 0)
	})
}

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
				IsRequired:  random.Bool(t),
			},
			{
				Type:        model.FreeNumberQuestion,
				Title:       random.AlphaNumericString(t, 20),
				Description: random.PtrOrNil(t, random.AlphaNumericString(t, 100)),
				IsPublic:    random.Bool(t),
				IsOpen:      random.Bool(t),
				IsRequired:  random.Bool(t),
			},
			{
				Type:        model.SingleChoiceQuestion,
				Title:       random.AlphaNumericString(t, 20),
				Description: random.PtrOrNil(t, random.AlphaNumericString(t, 100)),
				IsPublic:    random.Bool(t),
				IsOpen:      random.Bool(t),
				IsRequired:  random.Bool(t),
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
				IsRequired:  random.Bool(t),
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
			assert.Equal(t, question.IsRequired, createdQuestion.IsRequired)
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

func TestUpdateQuestionGroup(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		questionGroup := mustCreateQuestionGroup(t, r, camp.ID)
		newQuestionGroup := model.QuestionGroup{
			Name:        random.AlphaNumericString(t, 20),
			Description: random.PtrOrNil(t, random.AlphaNumericString(t, 100)),
			Due:         random.Time(t),
		}

		err := r.UpdateQuestionGroup(t.Context(), questionGroup.ID, newQuestionGroup)

		assert.NoError(t, err)

		updatedQuestionGroup, err := r.GetQuestionGroup(questionGroup.ID)

		assert.NoError(t, err)
		assert.Equal(t, newQuestionGroup.Name, updatedQuestionGroup.Name)
		assert.Equal(t, newQuestionGroup.Description, updatedQuestionGroup.Description)
		assert.WithinDuration(t, newQuestionGroup.Due, updatedQuestionGroup.Due, time.Second)
		assert.Equal(t, camp.ID, updatedQuestionGroup.CampID)
	})
}

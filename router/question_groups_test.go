package router

import (
	"net/http"
	"testing"
	"time"

	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestGetQuestionGroups(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)

		questionGroup1 := model.QuestionGroup{
			Model: gorm.Model{
				ID: uint(random.PositiveInt(t)),
			},
			Name:        random.AlphaNumericString(t, 20),
			Description: random.PtrOrNil(t, random.AlphaNumericString(t, 100)),
			Due:         random.Time(t),
			Questions: []model.Question{
				{
					Model: gorm.Model{
						ID: uint(random.PositiveInt(t)),
					},
					Type:        model.FreeTextQuestion,
					Title:       random.AlphaNumericString(t, 30),
					Description: random.PtrOrNil(t, random.AlphaNumericString(t, 100)),
					IsPublic:    random.Bool(t),
					IsOpen:      random.Bool(t),
					Options: []model.Option{
						{
							Model: gorm.Model{
								ID: uint(random.PositiveInt(t)),
							},
							Content: random.AlphaNumericString(t, 20),
						},
					},
				},
			},
		}

		questionGroup2 := model.QuestionGroup{
			Model: gorm.Model{
				ID: uint(random.PositiveInt(t)),
			},
			Name:        random.AlphaNumericString(t, 20),
			Description: random.PtrOrNil(t, random.AlphaNumericString(t, 100)),
			Due:         random.Time(t),
			CampID:      uint(campID),
		}

		h.repo.MockQuestionGroupRepository.EXPECT().GetQuestionGroups(gomock.Any(), uint(campID)).Return([]model.QuestionGroup{questionGroup1, questionGroup2}, nil).Times(1)

		res := h.expect.GET("/api/camps/{campId}/question-groups", campID).
			Expect().
			Status(http.StatusOK).JSON().Array()

		res.Length().IsEqual(2)

		res1 := res.Value(0).Object()

		res1.Keys().ContainsAll("id", "name", "due")
		res1.Value("id").Number().IsEqual(questionGroup1.ID)
		res1.Value("name").String().IsEqual(questionGroup1.Name)

		if questionGroup1.Description != nil {
			res1.Value("description").String().IsEqual(*questionGroup1.Description)
		} else {
			res1.Keys().NotContainsAny("description")
		}

		res1.Value("due").String().IsEqual(questionGroup1.Due.Format(time.DateOnly))

		questions := res1.Value("questions").Array()

		questions.Length().IsEqual(1)

		question := questions.Value(0).Object()

		question.Keys().ContainsAll("id", "type", "title", "isPublic", "isOpen", "options")
		question.Value("id").Number().IsEqual(questionGroup1.Questions[0].ID)
		question.Value("type").String().IsEqual(string(questionGroup1.Questions[0].Type))
		question.Value("title").String().IsEqual(questionGroup1.Questions[0].Title)

		if questionGroup1.Questions[0].Description != nil {
			question.Value("description").String().IsEqual(*questionGroup1.Questions[0].Description)
		} else {
			question.Keys().NotContainsAny("description")
		}

		question.Value("isPublic").Boolean().IsEqual(questionGroup1.Questions[0].IsPublic)
		question.Value("isOpen").Boolean().IsEqual(questionGroup1.Questions[0].IsOpen)

		options := question.Value("options").Array()

		options.Length().IsEqual(1)

		option := options.Value(0).Object()

		option.Keys().ContainsAll("id", "content")

		option.Value("id").Number().IsEqual(questionGroup1.Questions[0].Options[0].ID)
		option.Value("content").String().IsEqual(questionGroup1.Questions[0].Options[0].Content)

		res2 := res.Value(1).Object()

		res2.Keys().ContainsAll("id", "name", "due")
		res2.Value("id").Number().IsEqual(questionGroup2.ID)
		res2.Value("name").String().IsEqual(questionGroup2.Name)

		if questionGroup2.Description != nil {
			res2.Value("description").String().IsEqual(*questionGroup2.Description)
		} else {
			res2.Keys().NotContainsAny("description")
		}

		res2.Value("due").String().IsEqual(questionGroup2.Due.Format(time.DateOnly))
	})
}

func TestAdminPostQuestionGroup(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		freeTextQuestion := api.FreeTextQuestionRequest{
			Title:       random.AlphaNumericString(t, 30),
			Description: random.PtrOrNil(t, random.AlphaNumericString(t, 100)),
			Type:        api.FreeTextQuestionRequestTypeFreeText,
			IsPublic:    random.Bool(t),
			IsOpen:      random.Bool(t),
		}

		freeNumberQuestion := api.FreeNumberQuestionRequest{
			Title:       random.AlphaNumericString(t, 30),
			Description: random.PtrOrNil(t, random.AlphaNumericString(t, 100)),
			Type:        api.FreeNumberQuestionRequestTypeFreeNumber,
			IsPublic:    random.Bool(t),
			IsOpen:      random.Bool(t),
		}

		singleChoiceQuestion := api.PostSingleChoiceQuestionRequest{
			Title:       random.AlphaNumericString(t, 30),
			Description: random.PtrOrNil(t, random.AlphaNumericString(t, 100)),
			Type:        api.PostSingleChoiceQuestionRequestTypeSingle,
			IsPublic:    random.Bool(t),
			IsOpen:      random.Bool(t),
			Options: []api.PostOptionRequest{
				{
					Content: random.AlphaNumericString(t, 20),
				},
				{
					Content: random.AlphaNumericString(t, 20),
				},
			},
		}

		multipleChoiceQuestion := api.PostMultipleChoiceQuestionRequest{
			Title:       random.AlphaNumericString(t, 30),
			Description: random.PtrOrNil(t, random.AlphaNumericString(t, 100)),
			Type:        api.PostMultipleChoiceQuestionRequestTypeMultiple,
			IsPublic:    random.Bool(t),
			IsOpen:      random.Bool(t),
			Options: []api.PostOptionRequest{
				{
					Content: random.AlphaNumericString(t, 20),
				},
				{
					Content: random.AlphaNumericString(t, 20),
				},
				{
					Content: random.AlphaNumericString(t, 20),
				},
			},
		}

		questions := make([]api.PostQuestionRequest, 4)

		var freeTextReq api.PostQuestionRequest
		err := freeTextReq.FromFreeTextQuestionRequest(freeTextQuestion)
		require.NoError(t, err)
		questions[0] = freeTextReq

		var freeNumberReq api.PostQuestionRequest
		err = freeNumberReq.FromFreeNumberQuestionRequest(freeNumberQuestion)
		require.NoError(t, err)
		questions[1] = freeNumberReq

		var singleChoiceReq api.PostQuestionRequest
		err = singleChoiceReq.FromPostSingleChoiceQuestionRequest(singleChoiceQuestion)
		require.NoError(t, err)
		questions[2] = singleChoiceReq

		var multipleChoiceReq api.PostQuestionRequest
		err = multipleChoiceReq.FromPostMultipleChoiceQuestionRequest(multipleChoiceQuestion)
		require.NoError(t, err)
		questions[3] = multipleChoiceReq

		due := random.Time(t)
		req := api.AdminPostQuestionGroupJSONRequestBody{
			Name:        random.AlphaNumericString(t, 20),
			Description: random.PtrOrNil(t, random.AlphaNumericString(t, 100)),
			Due:         types.Date{Time: due},
			Questions:   questions,
		}
		username := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().GetOrCreateUser(gomock.Any(), username).Return(&model.User{IsStaff: true}, nil)
		h.repo.MockQuestionGroupRepository.EXPECT().CreateQuestionGroup(gomock.Any()).Return(nil)

		res := h.expect.POST("/api/admin/camps/1/question-groups").
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusCreated).JSON().Object()

		res.Keys().ContainsAll("id", "name", "due", "questions")
		res.Value("name").String().IsEqual(req.Name)

		if req.Description != nil {
			res.Value("description").String().IsEqual(*req.Description)
		}

		res.Value("due").String().IsEqual(req.Due.Format(time.DateOnly))

		questionsArray := res.Value("questions").Array()
		questionsArray.Length().IsEqual(4)

		// Verify FreeTextQuestion
		freeTextRes := questionsArray.Value(0).Object()
		freeTextRes.Value("title").String().IsEqual(freeTextQuestion.Title)

		if freeTextQuestion.Description != nil {
			freeTextRes.Value("description").String().IsEqual(*freeTextQuestion.Description)
		}

		freeTextRes.Value("type").String().IsEqual(string(freeTextQuestion.Type))
		freeTextRes.Value("isPublic").Boolean().IsEqual(freeTextQuestion.IsPublic)
		freeTextRes.Value("isOpen").Boolean().IsEqual(freeTextQuestion.IsOpen)

		// Verify FreeNumberQuestion
		freeNumberRes := questionsArray.Value(1).Object()
		freeNumberRes.Value("title").String().IsEqual(freeNumberQuestion.Title)

		if freeNumberQuestion.Description != nil {
			freeNumberRes.Value("description").String().IsEqual(*freeNumberQuestion.Description)
		}

		freeNumberRes.Value("type").String().IsEqual(string(freeNumberQuestion.Type))
		freeNumberRes.Value("isPublic").Boolean().IsEqual(freeNumberQuestion.IsPublic)
		freeNumberRes.Value("isOpen").Boolean().IsEqual(freeNumberQuestion.IsOpen)

		// Verify SingleChoiceQuestion
		singleChoiceRes := questionsArray.Value(2).Object()
		singleChoiceRes.Value("title").String().IsEqual(singleChoiceQuestion.Title)

		if singleChoiceQuestion.Description != nil {
			singleChoiceRes.Value("description").String().IsEqual(*singleChoiceQuestion.Description)
		}

		singleChoiceRes.Value("type").String().IsEqual(string(singleChoiceQuestion.Type))
		singleChoiceRes.Value("isPublic").Boolean().IsEqual(singleChoiceQuestion.IsPublic)
		singleChoiceRes.Value("isOpen").Boolean().IsEqual(singleChoiceQuestion.IsOpen)

		// Verify options for SingleChoiceQuestion
		singleChoiceOptions := singleChoiceRes.Value("options").Array()
		singleChoiceOptions.Length().IsEqual(2)
		singleChoiceOptions.Value(0).Object().Value("content").String().IsEqual(singleChoiceQuestion.Options[0].Content)
		singleChoiceOptions.Value(1).Object().Value("content").String().IsEqual(singleChoiceQuestion.Options[1].Content)

		// Verify MultipleChoiceQuestion
		multipleChoiceRes := questionsArray.Value(3).Object()
		multipleChoiceRes.Value("title").String().IsEqual(multipleChoiceQuestion.Title)

		if multipleChoiceQuestion.Description != nil {
			multipleChoiceRes.Value("description").String().IsEqual(*multipleChoiceQuestion.Description)
		}

		multipleChoiceRes.Value("type").String().IsEqual(string(multipleChoiceQuestion.Type))
		multipleChoiceRes.Value("isPublic").Boolean().IsEqual(multipleChoiceQuestion.IsPublic)
		multipleChoiceRes.Value("isOpen").Boolean().IsEqual(multipleChoiceQuestion.IsOpen)

		// Verify options for MultipleChoiceQuestion
		multipleChoiceOptions := multipleChoiceRes.Value("options").Array()
		multipleChoiceOptions.Length().IsEqual(3)
		multipleChoiceOptions.Value(0).Object().Value("content").String().IsEqual(multipleChoiceQuestion.Options[0].Content)
		multipleChoiceOptions.Value(1).Object().Value("content").String().IsEqual(multipleChoiceQuestion.Options[1].Content)
		multipleChoiceOptions.Value(2).Object().Value("content").String().IsEqual(multipleChoiceQuestion.Options[2].Content)
	})
}

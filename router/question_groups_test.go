package router

import (
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/traP-jp/rucQ/backend/api"
	"github.com/traP-jp/rucQ/backend/model"
	"github.com/traP-jp/rucQ/backend/testutil/random"
)

func TestAdminPostQuestionGroup(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		freeTextQuestion := api.FreeTextQuestionRequest{
			Title:       random.AlphaNumericString(t, 30),
			Description: random.Nilable(t, random.AlphaNumericString(t, 100)),
			Type:        api.FreeTextQuestionRequestTypeFreeText,
			IsPublic:    random.Bool(t),
			IsOpen:      random.Bool(t),
		}

		singleChoiceQuestion := api.SingleChoiceQuestionRequest{
			Title:       random.AlphaNumericString(t, 30),
			Description: random.Nilable(t, random.AlphaNumericString(t, 100)),
			Type:        api.SingleChoiceQuestionRequestTypeSingle,
			IsPublic:    random.Bool(t),
			IsOpen:      random.Bool(t),
			Options: []api.OptionRequest{
				{
					Content: random.AlphaNumericString(t, 20),
				},
				{
					Content: random.AlphaNumericString(t, 20),
				},
			},
		}

		multipleChoiceQuestion := api.MultipleChoiceQuestionRequest{
			Title:       random.AlphaNumericString(t, 30),
			Description: random.Nilable(t, random.AlphaNumericString(t, 100)),
			Type:        api.MultipleChoiceQuestionRequestTypeMultiple,
			IsPublic:    random.Bool(t),
			IsOpen:      random.Bool(t),
			Options: []api.OptionRequest{
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

		questions := make([]api.QuestionRequest, 3)

		var freeTextReq api.QuestionRequest
		freeTextReq.FromFreeTextQuestionRequest(freeTextQuestion)
		questions[0] = freeTextReq

		var singleChoiceReq api.QuestionRequest
		singleChoiceReq.FromSingleChoiceQuestionRequest(singleChoiceQuestion)
		questions[1] = singleChoiceReq

		var multipleChoiceReq api.QuestionRequest
		multipleChoiceReq.FromMultipleChoiceQuestionRequest(multipleChoiceQuestion)
		questions[2] = multipleChoiceReq

		due := random.Time(t)
		req := api.AdminPostQuestionGroupJSONRequestBody{
			Name:        random.AlphaNumericString(t, 20),
			Description: random.Nilable(t, random.AlphaNumericString(t, 100)),
			Due:         due,
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

		res.Keys().ContainsOnly("id", "name", "description", "due", "questions")
		res.Value("name").String().IsEqual(req.Name)

		if req.Description == nil {
			res.Value("description").IsNull()
		} else {
			res.Value("description").String().IsEqual(*req.Description)
		}

		res.Value("due").String().AsDateTime(time.RFC3339).InRange(req.Due.Add(-time.Second), req.Due.Add(time.Second))

		questionsArray := res.Value("questions").Array()
		questionsArray.Length().IsEqual(3)

		// Verify FreeTextQuestion
		freeTextRes := questionsArray.Value(0).Object()
		freeTextRes.Value("title").String().IsEqual(freeTextQuestion.Title)

		if freeTextQuestion.Description == nil {
			freeTextRes.Value("description").IsNull()
		} else {
			freeTextRes.Value("description").String().IsEqual(*freeTextQuestion.Description)
		}

		freeTextRes.Value("type").String().IsEqual(string(freeTextQuestion.Type))
		freeTextRes.Value("isPublic").Boolean().IsEqual(freeTextQuestion.IsPublic)
		freeTextRes.Value("isOpen").Boolean().IsEqual(freeTextQuestion.IsOpen)

		// Verify SingleChoiceQuestion
		singleChoiceRes := questionsArray.Value(1).Object()
		singleChoiceRes.Value("title").String().IsEqual(singleChoiceQuestion.Title)

		if singleChoiceQuestion.Description == nil {
			singleChoiceRes.Value("description").IsNull()
		} else {
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
		multipleChoiceRes := questionsArray.Value(2).Object()
		multipleChoiceRes.Value("title").String().IsEqual(multipleChoiceQuestion.Title)

		if multipleChoiceQuestion.Description == nil {
			multipleChoiceRes.Value("description").IsNull()
		} else {
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

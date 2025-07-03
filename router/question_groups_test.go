package router

import (
	"net/http"
	"testing"
	"time"

	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

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

		res.Keys().ContainsOnly("id", "name", "description", "due", "questions")
		res.Value("name").String().IsEqual(req.Name)

		if req.Description == nil {
			res.Value("description").IsNull()
		} else {
			res.Value("description").String().IsEqual(*req.Description)
		}

		res.Value("due").String().IsEqual(req.Due.Format(time.DateOnly))

		questionsArray := res.Value("questions").Array()
		questionsArray.Length().IsEqual(4)

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

		// Verify FreeNumberQuestion
		freeNumberRes := questionsArray.Value(1).Object()
		freeNumberRes.Value("title").String().IsEqual(freeNumberQuestion.Title)

		if freeNumberQuestion.Description == nil {
			freeNumberRes.Value("description").IsNull()
		} else {
			freeNumberRes.Value("description").String().IsEqual(*freeNumberQuestion.Description)
		}

		freeNumberRes.Value("type").String().IsEqual(string(freeNumberQuestion.Type))
		freeNumberRes.Value("isPublic").Boolean().IsEqual(freeNumberQuestion.IsPublic)
		freeNumberRes.Value("isOpen").Boolean().IsEqual(freeNumberQuestion.IsOpen)

		// Verify SingleChoiceQuestion
		singleChoiceRes := questionsArray.Value(2).Object()
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
		multipleChoiceRes := questionsArray.Value(3).Object()
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

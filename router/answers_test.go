package router

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestPostAnswers(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userID := random.AlphaNumericString(t, 32)

		freeTextAnswer := api.FreeTextAnswerRequest{
			Type:       api.FreeTextAnswerRequestTypeFreeText,
			QuestionId: random.PositiveInt(t),
			Content:    random.AlphaNumericString(t, 50),
		}
		var freeTextReq api.AnswerRequest
		err := freeTextReq.FromFreeTextAnswerRequest(freeTextAnswer)
		require.NoError(t, err)

		freeNumberAnswer := api.FreeNumberAnswerRequest{
			Type:       api.FreeNumberAnswerRequestTypeFreeNumber,
			QuestionId: random.PositiveInt(t),
			Content:    random.Float32(t),
		}
		var freeNumberReq api.AnswerRequest
		err = freeNumberReq.FromFreeNumberAnswerRequest(freeNumberAnswer)
		require.NoError(t, err)

		singleChoiceAnswer := api.SingleChoiceAnswerRequest{
			Type:       api.SingleChoiceAnswerRequestTypeSingle,
			QuestionId: random.PositiveInt(t),
			OptionId:   random.PositiveInt(t),
		}
		var singleChoiceReq api.AnswerRequest
		err = singleChoiceReq.FromSingleChoiceAnswerRequest(singleChoiceAnswer)
		require.NoError(t, err)

		multipleChoiceAnswer := api.MultipleChoiceAnswerRequest{
			Type:       api.MultipleChoiceAnswerRequestTypeMultiple,
			QuestionId: random.PositiveInt(t),
			OptionIds:  []int{random.PositiveInt(t), random.PositiveInt(t)},
		}
		var multipleChoiceReq api.AnswerRequest
		err = multipleChoiceReq.FromMultipleChoiceAnswerRequest(multipleChoiceAnswer)
		require.NoError(t, err)

		questionGroupID := random.PositiveInt(t)
		req := []api.AnswerRequest{
			freeTextReq,
			freeNumberReq,
			singleChoiceReq,
			multipleChoiceReq,
		}

		h.repo.MockAnswerRepository.EXPECT().
			CreateAnswers(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		res := h.expect.POST("/api/question-groups/{questionGroupId}/answers", questionGroupID).
			WithJSON(api.PostAnswersJSONRequestBody(req)).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusCreated).JSON().Array()

		res.Length().IsEqual(len(req))

		freeTextRes := res.Value(0).Object()

		freeTextRes.Keys().ContainsOnly("id", "type", "userId", "questionId", "content")
		freeTextRes.Value("type").String().IsEqual(string(freeTextAnswer.Type))
		freeTextRes.Value("userId").String().IsEqual(userID)
		freeTextRes.Value("questionId").Number().IsEqual(freeTextAnswer.QuestionId)
		freeTextRes.Value("content").String().IsEqual(freeTextAnswer.Content)

		freeNumberRes := res.Value(1).Object()

		freeNumberRes.Keys().ContainsOnly("id", "type", "userId", "questionId", "content")
		freeNumberRes.Value("type").String().IsEqual(string(freeNumberAnswer.Type))
		freeNumberRes.Value("userId").String().IsEqual(userID)
		freeNumberRes.Value("questionId").Number().IsEqual(freeNumberAnswer.QuestionId)
		freeNumberRes.Value("content").
			Number().
			InRange(float64(freeNumberAnswer.Content)-0.0001, float64(freeNumberAnswer.Content)+0.0001)

		singleChoiceRes := res.Value(2).Object()
		singleChoiceRes.Keys().ContainsOnly("id", "type", "userId", "questionId", "selectedOption")
		singleChoiceRes.Value("type").String().IsEqual(string(singleChoiceAnswer.Type))
		singleChoiceRes.Value("questionId").Number().IsEqual(singleChoiceAnswer.QuestionId)

		singleChoiceSelectedOption := singleChoiceRes.Value("selectedOption").Object()

		singleChoiceSelectedOption.Keys().ContainsOnly("id", "content")
		singleChoiceSelectedOption.Value("id").Number().IsEqual(singleChoiceAnswer.OptionId)

		multipleChoiceRes := res.Value(3).Object()
		multipleChoiceRes.Keys().
			ContainsOnly("id", "type", "userId", "questionId", "selectedOptions")
		multipleChoiceRes.Value("type").String().IsEqual(string(multipleChoiceAnswer.Type))
		multipleChoiceRes.Value("userId").String().IsEqual(userID)
		multipleChoiceRes.Value("questionId").Number().IsEqual(multipleChoiceAnswer.QuestionId)

		multipleChoiceSelectedOptions := multipleChoiceRes.Value("selectedOptions").Array()

		multipleChoiceSelectedOptions.Length().IsEqual(len(multipleChoiceAnswer.OptionIds))

		for i, optionID := range multipleChoiceAnswer.OptionIds {
			option := multipleChoiceSelectedOptions.Value(i).Object()

			option.Keys().ContainsOnly("id", "content")
			option.Value("id").Number().IsEqual(optionID)
		}
	})
}

func TestGetMyAnswers(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userID := random.AlphaNumericString(t, 32)
		questionGroupID := random.PositiveInt(t)

		freeTextContent := random.AlphaNumericString(t, 50)
		freeNumberContent := random.Float64(t)
		answers := []model.Answer{
			{
				QuestionID:      uint(random.PositiveInt(t)),
				UserID:          userID,
				Type:            model.FreeTextQuestion,
				FreeTextContent: &freeTextContent,
			},
			{
				QuestionID:        uint(random.PositiveInt(t)),
				UserID:            userID,
				Type:              model.FreeNumberQuestion,
				FreeNumberContent: &freeNumberContent,
			},
		}

		h.repo.MockAnswerRepository.EXPECT().
			GetAnswersByUserAndQuestionGroup(gomock.Any(), userID, uint(questionGroupID)).
			Return(answers, nil).
			Times(1)

		res := h.expect.GET("/api/me/question-groups/{questionGroupId}/answers", questionGroupID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusOK).JSON().Array()

		res.Length().IsEqual(len(answers))

		freeTextRes := res.Value(0).Object()
		freeTextRes.Keys().ContainsOnly("id", "type", "userId", "questionId", "content")
		freeTextRes.Value("type").String().IsEqual(string(model.FreeTextQuestion))
		freeTextRes.Value("userId").String().IsEqual(userID)
		freeTextRes.Value("questionId").Number().IsEqual(answers[0].QuestionID)
		freeTextRes.Value("content").String().IsEqual(freeTextContent)

		freeNumberRes := res.Value(1).Object()
		freeNumberRes.Keys().ContainsOnly("id", "type", "userId", "questionId", "content")
		freeNumberRes.Value("type").String().IsEqual(string(model.FreeNumberQuestion))
		freeNumberRes.Value("userId").String().IsEqual(userID)
		freeNumberRes.Value("questionId").Number().IsEqual(answers[1].QuestionID)
		freeNumberRes.Value("content").
			Number().
			InRange(freeNumberContent-0.0001, freeNumberContent+0.0001)
	})
}

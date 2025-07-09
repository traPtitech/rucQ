package router

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

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

func TestPutAnswer(t *testing.T) {
	t.Parallel()

	t.Run("Success with FreeText", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userID := random.AlphaNumericString(t, 32)
		answerID := uint(random.PositiveInt(t))

		freeTextAnswer := api.FreeTextAnswerRequest{
			Type:       api.FreeTextAnswerRequestTypeFreeText,
			QuestionId: random.PositiveInt(t),
			Content:    random.AlphaNumericString(t, 50),
		}
		var req api.AnswerRequest
		err := req.FromFreeTextAnswerRequest(freeTextAnswer)
		require.NoError(t, err)

		h.repo.MockAnswerRepository.EXPECT().
			UpdateAnswer(gomock.Any(), answerID, gomock.Any()).
			DoAndReturn(func(_ any, id uint, answer *model.Answer) error {
				answer.ID = id

				return nil
			}).
			Times(1)

		res := h.expect.PUT("/api/answers/{answerId}", answerID).
			WithJSON(api.PutAnswerJSONRequestBody(req)).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusOK).JSON().Object()

		res.Keys().ContainsOnly("id", "type", "userId", "questionId", "content")
		res.Value("id").Number().IsEqual(answerID)
		res.Value("type").String().IsEqual(string(freeTextAnswer.Type))
		res.Value("userId").String().IsEqual(userID)
		res.Value("questionId").Number().IsEqual(freeTextAnswer.QuestionId)
		res.Value("content").String().IsEqual(freeTextAnswer.Content)
	})

	t.Run("Success with FreeNumber", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userID := random.AlphaNumericString(t, 32)
		answerID := uint(random.PositiveInt(t))

		freeNumberAnswer := api.FreeNumberAnswerRequest{
			Type:       api.FreeNumberAnswerRequestTypeFreeNumber,
			QuestionId: random.PositiveInt(t),
			Content:    random.Float32(t),
		}
		var req api.AnswerRequest
		err := req.FromFreeNumberAnswerRequest(freeNumberAnswer)
		require.NoError(t, err)

		h.repo.MockAnswerRepository.EXPECT().
			UpdateAnswer(gomock.Any(), answerID, gomock.Any()).
			DoAndReturn(func(_ any, id uint, answer *model.Answer) error {
				answer.ID = id

				return nil
			}).
			Times(1)

		res := h.expect.PUT("/api/answers/{answerId}", answerID).
			WithJSON(api.PutAnswerJSONRequestBody(req)).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusOK).JSON().Object()

		res.Keys().ContainsOnly("id", "type", "userId", "questionId", "content")
		res.Value("id").Number().IsEqual(answerID)
		res.Value("type").String().IsEqual(string(freeNumberAnswer.Type))
		res.Value("userId").String().IsEqual(userID)
		res.Value("questionId").Number().IsEqual(freeNumberAnswer.QuestionId)
		res.Value("content").
			Number().
			InRange(float64(freeNumberAnswer.Content)-0.0001, float64(freeNumberAnswer.Content)+0.0001)
	})

	t.Run("Success with SingleChoice", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userID := random.AlphaNumericString(t, 32)
		answerID := uint(random.PositiveInt(t))

		optionID := random.PositiveInt(t)
		singleChoiceAnswer := api.SingleChoiceAnswerRequest{
			Type:       api.SingleChoiceAnswerRequestTypeSingle,
			QuestionId: random.PositiveInt(t),
			OptionId:   optionID,
		}
		var req api.AnswerRequest
		err := req.FromSingleChoiceAnswerRequest(singleChoiceAnswer)
		require.NoError(t, err)

		h.repo.MockAnswerRepository.EXPECT().
			UpdateAnswer(gomock.Any(), answerID, gomock.Any()).
			DoAndReturn(func(_ any, id uint, answer *model.Answer) error {
				answer.ID = id
				answer.SelectedOptions = []model.Option{
					{
						Model: gorm.Model{
							ID: uint(optionID),
						},
						Content: random.AlphaNumericString(t, 20),
					},
				}

				return nil
			}).
			Times(1)

		res := h.expect.PUT("/api/answers/{answerId}", answerID).
			WithJSON(api.PutAnswerJSONRequestBody(req)).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusOK).JSON().Object()

		res.Keys().ContainsOnly("id", "type", "userId", "questionId", "selectedOption")
		res.Value("id").Number().IsEqual(answerID)
		res.Value("type").String().IsEqual(string(singleChoiceAnswer.Type))
		res.Value("userId").String().IsEqual(userID)
		res.Value("questionId").Number().IsEqual(singleChoiceAnswer.QuestionId)

		selectedOption := res.Value("selectedOption").Object()
		selectedOption.Keys().ContainsOnly("id", "content")
		selectedOption.Value("id").Number().IsEqual(singleChoiceAnswer.OptionId)
		selectedOption.Value("content").String().NotEmpty()
	})

	t.Run("Success with MultipleChoice", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userID := random.AlphaNumericString(t, 32)
		answerID := uint(random.PositiveInt(t))

		optionID1 := random.PositiveInt(t)
		optionID2 := random.PositiveInt(t)
		multipleChoiceAnswer := api.MultipleChoiceAnswerRequest{
			Type:       api.MultipleChoiceAnswerRequestTypeMultiple,
			QuestionId: random.PositiveInt(t),
			OptionIds:  []int{optionID1, optionID2},
		}
		var req api.AnswerRequest
		err := req.FromMultipleChoiceAnswerRequest(multipleChoiceAnswer)
		require.NoError(t, err)

		h.repo.MockAnswerRepository.EXPECT().
			UpdateAnswer(gomock.Any(), answerID, gomock.Any()).
			DoAndReturn(func(_ any, id uint, answer *model.Answer) error {
				answer.ID = id
				answer.SelectedOptions = []model.Option{
					{
						Model: gorm.Model{
							ID: uint(optionID1),
						},
						Content: random.AlphaNumericString(t, 20),
					},
					{
						Model: gorm.Model{
							ID: uint(optionID2),
						},
						Content: random.AlphaNumericString(t, 20),
					},
				}

				return nil
			}).
			Times(1)

		res := h.expect.PUT("/api/answers/{answerId}", answerID).
			WithJSON(api.PutAnswerJSONRequestBody(req)).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusOK).JSON().Object()

		res.Keys().ContainsOnly("id", "type", "userId", "questionId", "selectedOptions")
		res.Value("id").Number().IsEqual(answerID)
		res.Value("type").String().IsEqual(string(multipleChoiceAnswer.Type))
		res.Value("userId").String().IsEqual(userID)
		res.Value("questionId").Number().IsEqual(multipleChoiceAnswer.QuestionId)

		selectedOptions := res.Value("selectedOptions").Array()
		selectedOptions.Length().IsEqual(len(multipleChoiceAnswer.OptionIds))

		for i, optionID := range multipleChoiceAnswer.OptionIds {
			option := selectedOptions.Value(i).Object()
			option.Keys().ContainsOnly("id", "content")
			option.Value("id").Number().IsEqual(optionID)
			option.Value("content").String().NotEmpty()
		}
	})
}

func TestAdminGetAnswers(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userID := random.AlphaNumericString(t, 32)
		questionID := random.PositiveInt(t)

		freeTextContent := random.AlphaNumericString(t, 50)
		freeNumberContent := random.Float64(t)
		optionID1 := random.PositiveInt(t)
		optionID2 := random.PositiveInt(t)

		answers := []model.Answer{
			{
				Model: gorm.Model{
					ID: uint(random.PositiveInt(t)),
				},
				QuestionID:      uint(questionID),
				UserID:          userID,
				Type:            model.FreeTextQuestion,
				FreeTextContent: &freeTextContent,
			},
			{
				Model: gorm.Model{
					ID: uint(random.PositiveInt(t)),
				},
				QuestionID:        uint(questionID),
				UserID:            userID,
				Type:              model.FreeNumberQuestion,
				FreeNumberContent: &freeNumberContent,
			},
			{
				Model: gorm.Model{
					ID: uint(random.PositiveInt(t)),
				},
				QuestionID: uint(questionID),
				UserID:     userID,
				Type:       model.SingleChoiceQuestion,
				SelectedOptions: []model.Option{
					{
						Model: gorm.Model{
							ID: uint(optionID1),
						},
						Content: random.AlphaNumericString(t, 20),
					},
				},
			},
			{
				Model: gorm.Model{
					ID: uint(random.PositiveInt(t)),
				},
				QuestionID: uint(questionID),
				UserID:     userID,
				Type:       model.MultipleChoiceQuestion,
				SelectedOptions: []model.Option{
					{
						Model: gorm.Model{
							ID: uint(optionID1),
						},
						Content: random.AlphaNumericString(t, 20),
					},
					{
						Model: gorm.Model{
							ID: uint(optionID2),
						},
						Content: random.AlphaNumericString(t, 20),
					},
				},
			},
		}

		// スタッフユーザーを設定
		staffUser := &model.User{
			ID:      userID,
			IsStaff: true,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(staffUser, nil).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			GetAnswersByQuestionID(gomock.Any(), uint(questionID)).
			Return(answers, nil).
			Times(1)

		res := h.expect.GET("/api/admin/questions/{questionId}/answers", questionID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusOK).JSON().Array()

		res.Length().IsEqual(len(answers))

		for i, answer := range answers {
			answerObj := res.Value(i).Object()
			answerObj.Value("id").Number().IsEqual(answer.ID)
			answerObj.Value("userId").String().IsEqual(answer.UserID)
			answerObj.Value("questionId").Number().IsEqual(answer.QuestionID)

			switch answer.Type {
			case model.FreeTextQuestion:
				answerObj.Keys().ContainsOnly("id", "type", "userId", "questionId", "content")
				answerObj.Value("type").String().IsEqual("free_text")
				answerObj.Value("content").String().IsEqual(*answer.FreeTextContent)
			case model.FreeNumberQuestion:
				answerObj.Keys().ContainsOnly("id", "type", "userId", "questionId", "content")
				answerObj.Value("type").String().IsEqual("free_number")
				// float64からfloat32への変換で精度が変わるため、InRangeを使用
				expectedContent := float32(*answer.FreeNumberContent)
				answerObj.Value("content").Number().
					InRange(float64(expectedContent)-0.0001, float64(expectedContent)+0.0001)
			case model.SingleChoiceQuestion:
				answerObj.Keys().
					ContainsOnly("id", "type", "userId", "questionId", "selectedOption")
				answerObj.Value("type").String().IsEqual("single")
				selectedOption := answerObj.Value("selectedOption").Object()
				selectedOption.Value("id").Number().IsEqual(answer.SelectedOptions[0].ID)
				selectedOption.Value("content").String().IsEqual(answer.SelectedOptions[0].Content)
			case model.MultipleChoiceQuestion:
				answerObj.Keys().
					ContainsOnly("id", "type", "userId", "questionId", "selectedOptions")
				answerObj.Value("type").String().IsEqual("multiple")
				selectedOptions := answerObj.Value("selectedOptions").Array()
				selectedOptions.Length().IsEqual(2)
				selectedOptions.Value(0).
					Object().
					Value("id").
					Number().
					IsEqual(answer.SelectedOptions[0].ID)
				selectedOptions.Value(0).
					Object().
					Value("content").
					String().
					IsEqual(answer.SelectedOptions[0].Content)
				selectedOptions.Value(1).
					Object().
					Value("id").
					Number().
					IsEqual(answer.SelectedOptions[1].ID)
				selectedOptions.Value(1).
					Object().
					Value("content").
					String().
					IsEqual(answer.SelectedOptions[1].Content)
			}
		}
	})

	t.Run("Forbidden - Non-Staff User", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userID := random.AlphaNumericString(t, 32)
		questionID := random.PositiveInt(t)

		// 非スタッフユーザーを設定
		nonStaffUser := &model.User{
			ID:      userID,
			IsStaff: false,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(nonStaffUser, nil).
			Times(1)

		h.expect.GET("/api/admin/questions/{questionId}/answers", questionID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusForbidden).JSON().Object().
			Value("message").String().IsEqual("Forbidden")
	})

	t.Run("InternalServerError - User Repository Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userID := random.AlphaNumericString(t, 32)
		questionID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(nil, gorm.ErrRecordNotFound).
			Times(1)

		h.expect.GET("/api/admin/questions/{questionId}/answers", questionID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusInternalServerError).JSON().Object().
			Value("message").String().IsEqual("Internal server error")
	})

	t.Run("EmptyResult", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userID := random.AlphaNumericString(t, 32)
		questionID := random.PositiveInt(t)

		// スタッフユーザーを設定
		staffUser := &model.User{
			ID:      userID,
			IsStaff: true,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(staffUser, nil).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			GetAnswersByQuestionID(gomock.Any(), uint(questionID)).
			Return([]model.Answer{}, nil).
			Times(1)

		res := h.expect.GET("/api/admin/questions/{questionId}/answers", questionID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusOK).JSON().Array()

		res.Length().IsEqual(0)
	})

	t.Run("RepositoryError", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userID := random.AlphaNumericString(t, 32)
		questionID := random.PositiveInt(t)

		// スタッフユーザーを設定
		staffUser := &model.User{
			ID:      userID,
			IsStaff: true,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(staffUser, nil).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			GetAnswersByQuestionID(gomock.Any(), uint(questionID)).
			Return(nil, gorm.ErrRecordNotFound).
			Times(1)

		h.expect.GET("/api/admin/questions/{questionId}/answers", questionID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusInternalServerError).JSON().Object().
			Value("message").String().IsEqual("Internal server error")
	})
}

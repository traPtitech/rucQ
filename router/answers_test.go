package router

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
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
		questionGroupID := uint(random.PositiveInt(t))

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
			GetAnswers(gomock.Any(), repository.GetAnswersQuery{
				UserID:                &userID,
				QuestionGroupID:       &questionGroupID,
				IncludePrivateAnswers: true,
			}).
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

	t.Run("NotFound - Question group not found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userID := random.AlphaNumericString(t, 32)
		questionGroupID := uint(random.PositiveInt(t))

		h.repo.MockAnswerRepository.EXPECT().
			GetAnswers(gomock.Any(), repository.GetAnswersQuery{
				UserID:                &userID,
				QuestionGroupID:       &questionGroupID,
				IncludePrivateAnswers: true,
			}).
			Return(nil, model.ErrNotFound).
			Times(1)

		h.expect.GET("/api/me/question-groups/{questionGroupId}/answers", questionGroupID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusNotFound).JSON().Object().
			Value("message").String().IsEqual("Question group not found")
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
						Model:   gorm.Model{ID: uint(optionID1)},
						Content: random.AlphaNumericString(t, 20),
					},
					{
						Model:   gorm.Model{ID: uint(optionID2)},
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
		questionID := uint(random.PositiveInt(t))

		freeTextContent := random.AlphaNumericString(t, 50)
		freeNumberContent := random.Float64(t)
		optionID1 := random.PositiveInt(t)
		optionID2 := random.PositiveInt(t)

		answers := []model.Answer{
			{
				Model: gorm.Model{
					ID: uint(random.PositiveInt(t)),
				},
				QuestionID:      questionID,
				UserID:          userID,
				Type:            model.FreeTextQuestion,
				FreeTextContent: &freeTextContent,
			},
			{
				Model: gorm.Model{
					ID: uint(random.PositiveInt(t)),
				},
				QuestionID:        questionID,
				UserID:            userID,
				Type:              model.FreeNumberQuestion,
				FreeNumberContent: &freeNumberContent,
			},
			{
				Model: gorm.Model{
					ID: uint(random.PositiveInt(t)),
				},
				QuestionID: questionID,
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
				QuestionID: questionID,
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
			GetAnswers(gomock.Any(), repository.GetAnswersQuery{
				QuestionID:            &questionID,
				IncludePrivateAnswers: true,
			}).
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
		questionID := uint(random.PositiveInt(t))

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
			GetAnswers(gomock.Any(), repository.GetAnswersQuery{
				QuestionID:            &questionID,
				IncludePrivateAnswers: true,
			}).
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
		questionID := uint(random.PositiveInt(t))

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
			GetAnswers(gomock.Any(), repository.GetAnswersQuery{
				QuestionID:            &questionID,
				IncludePrivateAnswers: true,
			}).
			Return(nil, gorm.ErrRecordNotFound).
			Times(1)

		h.expect.GET("/api/admin/questions/{questionId}/answers", questionID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusInternalServerError).JSON().Object().
			Value("message").String().IsEqual("Internal server error")
	})

	t.Run("NotFound - Question Does Not Exist", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userID := random.AlphaNumericString(t, 32)
		questionID := uint(random.PositiveInt(t))
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
			GetAnswers(gomock.Any(), repository.GetAnswersQuery{
				QuestionID:            &questionID,
				IncludePrivateAnswers: true,
			}).
			Return(nil, model.ErrNotFound).
			Times(1)

		h.expect.GET("/api/admin/questions/{questionId}/answers", questionID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusNotFound).JSON().Object().
			Value("message").String().IsEqual("Question not found")
	})
}

func TestAdminPutAnswer(t *testing.T) {
	t.Parallel()

	t.Run("Success - Update FreeText Answer", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		wg.Add(1)

		h := setup(t)
		userID := random.AlphaNumericString(t, 32)
		answerID := random.PositiveInt(t)
		questionID := random.PositiveInt(t)

		// スタッフユーザーを設定
		staffUser := &model.User{
			ID:      userID,
			IsStaff: true,
		}

		updatedContent := random.AlphaNumericString(t, 50)

		// 変更前の回答 (GetAnswerByIDで返される)
		oldContent := random.AlphaNumericString(t, 50)
		oldAnswer := &model.Answer{
			Model:           gorm.Model{ID: uint(answerID)},
			UserID:          random.AlphaNumericString(t, 32), // 回答者のID
			QuestionID:      uint(questionID),
			Type:            model.FreeTextQuestion,
			FreeTextContent: &oldContent,
		}

		// 質問 (GetQuestionByIDで返される)
		question := &model.Question{
			Model: gorm.Model{ID: uint(questionID)},
			Title: random.AlphaNumericString(t, 20),
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(staffUser, nil).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			GetAnswerByID(gomock.Any(), uint(answerID)).
			Return(oldAnswer, nil).
			Times(1)

		h.repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(uint(questionID)).
			Return(question, nil).
			Times(1)

		expectedMessage := fmt.Sprintf(
			"@%sがアンケート「%s」のあなたの回答を変更しました。\n### 変更前\n```\n%s\n```\n### 変更後\n```\n%s\n```",
			userID,
			question.Title,
			*oldAnswer.FreeTextContent,
			updatedContent,
		)

		h.traqService.EXPECT().
			PostDirectMessage(gomock.Any(), oldAnswer.UserID, expectedMessage).
			DoAndReturn(func(_, _, _ any) error {
				defer wg.Done()

				return nil
			}).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			UpdateAnswer(gomock.Any(), uint(answerID), gomock.Any()).
			DoAndReturn(func(_ any, id uint, answer *model.Answer) error {
				answer.ID = id
				answer.UserID = oldAnswer.UserID
				answer.FreeTextContent = &updatedContent
				return nil
			}).
			Times(1)

		reqBody := api.FreeTextAnswerRequest{
			Type:       api.FreeTextAnswerRequestTypeFreeText,
			QuestionId: questionID,
			Content:    updatedContent,
		}

		var req api.AnswerRequest
		err := req.FromFreeTextAnswerRequest(reqBody)
		require.NoError(t, err)

		res := h.expect.PUT("/api/admin/answers/{answerId}", answerID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(req).
			Expect().
			Status(http.StatusOK).JSON().Object()

		waitWithTimeout(t, &wg, 2*time.Second)
		res.Value("id").Number().IsEqual(answerID)
		res.Value("type").String().IsEqual("free_text")
		res.Value("questionId").Number().IsEqual(questionID)
		res.Value("content").String().IsEqual(updatedContent)
		res.Keys().ContainsOnly("id", "type", "questionId", "content", "userId")
	})

	t.Run("Success - Update FreeNumber Answer", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		wg.Add(1)

		h := setup(t)
		userID := random.AlphaNumericString(t, 32)
		answerID := random.PositiveInt(t)
		questionID := random.PositiveInt(t)

		// スタッフユーザーを設定
		staffUser := &model.User{
			ID:      userID,
			IsStaff: true,
		}

		updatedContent := random.Float64(t)

		// 変更前の回答 (GetAnswerByIDで返される)
		oldContent := random.Float64(t)
		oldAnswer := &model.Answer{
			Model:             gorm.Model{ID: uint(answerID)},
			UserID:            random.AlphaNumericString(t, 32), // 回答者のID
			QuestionID:        uint(questionID),
			Type:              model.FreeNumberQuestion,
			FreeNumberContent: &oldContent,
		}

		// 質問 (GetQuestionByIDで返される)
		question := &model.Question{
			Model: gorm.Model{ID: uint(questionID)},
			Title: random.AlphaNumericString(t, 20),
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(staffUser, nil).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			GetAnswerByID(gomock.Any(), uint(answerID)).
			Return(oldAnswer, nil).
			Times(1)

		h.repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(uint(questionID)).
			Return(question, nil).
			Times(1)

		expectedMessage := fmt.Sprintf(
			"@%sがアンケート「%s」のあなたの回答を変更しました。\n### 変更前\n```\n%g\n```\n### 変更後\n```\n%g\n```",
			userID,
			question.Title,
			*oldAnswer.FreeNumberContent,
			updatedContent,
		)

		h.traqService.EXPECT().
			PostDirectMessage(gomock.Any(), oldAnswer.UserID, expectedMessage).
			DoAndReturn(func(_, _, _ any) error {
				defer wg.Done()

				return nil
			}).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			UpdateAnswer(gomock.Any(), uint(answerID), gomock.Any()).
			DoAndReturn(func(_ any, id uint, answer *model.Answer) error {
				answer.ID = id
				answer.UserID = oldAnswer.UserID
				answer.FreeNumberContent = &updatedContent
				return nil
			}).
			Times(1)

		reqBody := api.FreeNumberAnswerRequest{
			Type:       api.FreeNumberAnswerRequestTypeFreeNumber,
			QuestionId: questionID,
			Content:    float32(updatedContent),
		}

		var req api.AnswerRequest
		err := req.FromFreeNumberAnswerRequest(reqBody)
		require.NoError(t, err)

		res := h.expect.PUT("/api/admin/answers/{answerId}", answerID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(req).
			Expect().
			Status(http.StatusOK).JSON().Object()

		waitWithTimeout(t, &wg, 2*time.Second)
		res.Value("id").Number().IsEqual(answerID)
		res.Value("type").String().IsEqual("free_number")
		res.Value("questionId").Number().IsEqual(questionID)
		res.Value("content").
			Number().
			InRange(float64(updatedContent)-0.0001, float64(updatedContent)+0.0001)
		res.Keys().ContainsOnly("id", "type", "questionId", "content", "userId")
	})

	t.Run("Success - Update SingleChoice Answer", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		wg.Add(1)

		h := setup(t)
		userID := random.AlphaNumericString(t, 32)
		answerID := random.PositiveInt(t)
		questionID := random.PositiveInt(t)
		optionID := random.PositiveInt(t)

		// スタッフユーザーを設定
		staffUser := &model.User{
			ID:      userID,
			IsStaff: true,
		}

		option := model.Option{
			Model: gorm.Model{
				ID: uint(optionID),
			},
			Content: random.AlphaNumericString(t, 20),
		}

		// 変更前の回答 (GetAnswerByIDで返される)
		oldOption := model.Option{
			Model: gorm.Model{
				ID: uint(random.PositiveInt(t)),
			},
			Content: random.AlphaNumericString(t, 20),
		}
		oldAnswer := &model.Answer{
			Model:           gorm.Model{ID: uint(answerID)},
			UserID:          random.AlphaNumericString(t, 32), // 回答者のID
			QuestionID:      uint(questionID),
			Type:            model.SingleChoiceQuestion,
			SelectedOptions: []model.Option{oldOption},
		}

		// 質問 (GetQuestionByIDで返される)
		question := &model.Question{
			Model: gorm.Model{ID: uint(questionID)},
			Title: random.AlphaNumericString(t, 20),
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(staffUser, nil).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			GetAnswerByID(gomock.Any(), uint(answerID)).
			Return(oldAnswer, nil).
			Times(1)

		h.repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(uint(questionID)).
			Return(question, nil).
			Times(1)

		oldOptionContent := oldAnswer.SelectedOptions[0].Content
		newOptionContent := option.Content
		expectedMessage := fmt.Sprintf(
			"@%sがアンケート「%s」のあなたの回答を変更しました。\n### 変更前\n- %s\n### 変更後\n- %s",
			userID,
			question.Title,
			oldOptionContent,
			newOptionContent,
		)

		h.traqService.EXPECT().
			PostDirectMessage(gomock.Any(), oldAnswer.UserID, expectedMessage).
			DoAndReturn(func(_, _, _ any) error {
				defer wg.Done()

				return nil
			}).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			UpdateAnswer(gomock.Any(), uint(answerID), gomock.Any()).
			DoAndReturn(func(_ any, id uint, answer *model.Answer) error {
				answer.ID = id
				answer.UserID = oldAnswer.UserID
				answer.SelectedOptions = []model.Option{option}
				return nil
			}).
			Times(1)

		reqBody := api.SingleChoiceAnswerRequest{
			Type:       api.SingleChoiceAnswerRequestTypeSingle,
			QuestionId: questionID,
			OptionId:   optionID,
		}

		var req api.AnswerRequest
		err := req.FromSingleChoiceAnswerRequest(reqBody)
		require.NoError(t, err)

		res := h.expect.PUT("/api/admin/answers/{answerId}", answerID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(req).
			Expect().
			Status(http.StatusOK).JSON().Object()

		waitWithTimeout(t, &wg, 2*time.Second)
		res.Value("id").Number().IsEqual(answerID)
		res.Value("type").String().IsEqual("single")
		res.Value("questionId").Number().IsEqual(questionID)
		selectedOption := res.Value("selectedOption").Object()
		selectedOption.Value("id").Number().IsEqual(optionID)
		selectedOption.Value("content").String().IsEqual(option.Content)
		res.Keys().ContainsOnly("id", "type", "questionId", "selectedOption", "userId")
	})

	t.Run("Success - Update MultipleChoice Answer", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		wg.Add(1)

		h := setup(t)
		userID := random.AlphaNumericString(t, 32)
		answerID := random.PositiveInt(t)
		questionID := random.PositiveInt(t)
		optionID1 := random.PositiveInt(t)
		optionID2 := random.PositiveInt(t)

		// スタッフユーザーを設定
		staffUser := &model.User{
			ID:      userID,
			IsStaff: true,
		}

		options := []model.Option{
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

		// 変更前の回答 (GetAnswerByIDで返される)
		oldOptions := []model.Option{
			{
				Model: gorm.Model{
					ID: uint(random.PositiveInt(t)),
				},
				Content: random.AlphaNumericString(t, 20),
			},
		}
		oldAnswer := &model.Answer{
			Model:           gorm.Model{ID: uint(answerID)},
			UserID:          random.AlphaNumericString(t, 32), // 回答者のID
			QuestionID:      uint(questionID),
			Type:            model.MultipleChoiceQuestion,
			SelectedOptions: oldOptions,
		}

		// 質問 (GetQuestionByIDで返される)
		question := &model.Question{
			Model: gorm.Model{ID: uint(questionID)},
			Title: random.AlphaNumericString(t, 20),
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(staffUser, nil).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			GetAnswerByID(gomock.Any(), uint(answerID)).
			Return(oldAnswer, nil).
			Times(1)

		h.repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(uint(questionID)).
			Return(question, nil).
			Times(1)

		var oldOptionsString string

		for _, opt := range oldAnswer.SelectedOptions {
			oldOptionsString = fmt.Sprintf("%s- %s\n", oldOptionsString, opt.Content)
		}

		var newOptionsString string

		for _, opt := range options {
			newOptionsString = fmt.Sprintf("%s- %s\n", newOptionsString, opt.Content)
		}

		expectedMessage := fmt.Sprintf(
			"@%sがアンケート「%s」のあなたの回答を変更しました。\n### 変更前\n%s### 変更後\n%s",
			userID,
			question.Title,
			oldOptionsString,
			newOptionsString,
		)

		h.traqService.EXPECT().
			PostDirectMessage(gomock.Any(), oldAnswer.UserID, expectedMessage).
			DoAndReturn(func(_, _, _ any) error {
				defer wg.Done()

				return nil
			}).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			UpdateAnswer(gomock.Any(), uint(answerID), gomock.Any()).
			DoAndReturn(func(_ any, id uint, answer *model.Answer) error {
				answer.ID = id
				answer.UserID = oldAnswer.UserID
				answer.SelectedOptions = options
				return nil
			}).
			Times(1)

		reqBody := api.MultipleChoiceAnswerRequest{
			Type:       api.MultipleChoiceAnswerRequestTypeMultiple,
			QuestionId: questionID,
			OptionIds:  []int{optionID1, optionID2},
		}

		var req api.AnswerRequest
		err := req.FromMultipleChoiceAnswerRequest(reqBody)
		require.NoError(t, err)

		res := h.expect.PUT("/api/admin/answers/{answerId}", answerID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(req).
			Expect().
			Status(http.StatusOK).JSON().Object()

		waitWithTimeout(t, &wg, 2*time.Second)
		res.Keys().ContainsOnly("id", "type", "questionId", "selectedOptions", "userId")
		res.Value("id").Number().IsEqual(answerID)
		res.Value("type").String().IsEqual("multiple")
		res.Value("questionId").Number().IsEqual(questionID)
		selectedOptions := res.Value("selectedOptions").Array()
		selectedOptions.Length().IsEqual(2)

		for i, optionID := range []int{optionID1, optionID2} {
			option := selectedOptions.Value(i).Object()
			option.Keys().ContainsOnly("id", "content")
			option.Value("id").Number().IsEqual(optionID)
			option.Value("content").String().NotEmpty()
		}
	})
}

func TestAdminPostAnswer(t *testing.T) {
	t.Parallel()

	t.Run("Success - Create FreeText Answer", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		wg.Add(1)

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		questionID := random.PositiveInt(t)

		// 管理者ユーザーを設定
		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		// 対象ユーザーを設定
		targetUser := &model.User{
			ID:      targetUserID,
			IsStaff: false,
		}

		content := random.AlphaNumericString(t, 50)

		// 質問 (GetQuestionByIDで返される)
		question := &model.Question{
			Model: gorm.Model{ID: uint(questionID)},
			Title: random.AlphaNumericString(t, 20),
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), targetUserID).
			Return(targetUser, nil).
			Times(1)

		h.repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(uint(questionID)).
			Return(question, nil).
			Times(1)

		expectedMessage := fmt.Sprintf(
			"@%sがアンケート「%s」のあなたの回答を変更しました。\n### 変更前\n未回答\n### 変更後\n```\n%s\n```\n",
			adminUserID,
			question.Title,
			content,
		)

		h.traqService.EXPECT().
			PostDirectMessage(gomock.Any(), targetUserID, expectedMessage).
			DoAndReturn(func(_, _, _ any) error {
				defer wg.Done()

				return nil
			}).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			CreateAnswer(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ any, answer *model.Answer) error {
				answer.ID = uint(random.PositiveInt(t))
				// FreeTextContentを設定
				answer.FreeTextContent = &content

				return nil
			}).
			Times(1)

		reqBody := api.FreeTextAnswerRequest{
			Type:       api.FreeTextAnswerRequestTypeFreeText,
			QuestionId: questionID,
			Content:    content,
		}

		var req api.AnswerRequest
		err := req.FromFreeTextAnswerRequest(reqBody)
		require.NoError(t, err)

		res := h.expect.POST("/api/admin/users/{userId}/answers", targetUserID).
			WithHeader("X-Forwarded-User", adminUserID).
			WithJSON(req).
			Expect().
			Status(http.StatusCreated).JSON().Object()

		waitWithTimeout(t, &wg, 2*time.Second)

		res.Keys().ContainsOnly("id", "type", "userId", "questionId", "content")
		res.Value("type").String().IsEqual(string(api.FreeTextAnswerRequestTypeFreeText))
		res.Value("userId").String().IsEqual(targetUserID)
		res.Value("questionId").Number().IsEqual(questionID)
		res.Value("content").String().IsEqual(content)
	})

	t.Run("Success - Create FreeNumber Answer", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		wg.Add(1)

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		questionID := random.PositiveInt(t)

		// 管理者ユーザーを設定
		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		// 対象ユーザーを設定
		targetUser := &model.User{
			ID:      targetUserID,
			IsStaff: false,
		}

		content := random.Float32(t)
		contentFloat64 := float64(content)

		// 質問 (GetQuestionByIDで返される)
		question := &model.Question{
			Model: gorm.Model{ID: uint(questionID)},
			Title: random.AlphaNumericString(t, 20),
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), targetUserID).
			Return(targetUser, nil).
			Times(1)

		h.repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(uint(questionID)).
			Return(question, nil).
			Times(1)

		expectedMessage := fmt.Sprintf(
			"@%sがアンケート「%s」のあなたの回答を変更しました。\n### 変更前\n未回答\n### 変更後\n```\n%g\n```\n",
			adminUserID,
			question.Title,
			contentFloat64,
		)

		h.traqService.EXPECT().
			PostDirectMessage(gomock.Any(), targetUserID, expectedMessage).
			DoAndReturn(func(_, _, _ any) error {
				defer wg.Done()

				return nil
			}).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			CreateAnswer(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ any, answer *model.Answer) error {
				answer.ID = uint(random.PositiveInt(t))
				// FreeNumberContentを設定
				answer.FreeNumberContent = &contentFloat64

				return nil
			}).
			Times(1)

		reqBody := api.FreeNumberAnswerRequest{
			Type:       api.FreeNumberAnswerRequestTypeFreeNumber,
			QuestionId: questionID,
			Content:    content,
		}

		var req api.AnswerRequest
		err := req.FromFreeNumberAnswerRequest(reqBody)
		require.NoError(t, err)

		res := h.expect.POST("/api/admin/users/{userId}/answers", targetUserID).
			WithHeader("X-Forwarded-User", adminUserID).
			WithJSON(req).
			Expect().
			Status(http.StatusCreated).JSON().Object()

		waitWithTimeout(t, &wg, 2*time.Second)

		res.Keys().ContainsOnly("id", "type", "userId", "questionId", "content")
		res.Value("type").String().IsEqual(string(api.FreeNumberAnswerRequestTypeFreeNumber))
		res.Value("userId").String().IsEqual(targetUserID)
		res.Value("questionId").Number().IsEqual(questionID)
		res.Value("content").Number().InRange(float64(content)-0.0001, float64(content)+0.0001)
	})

	t.Run("Success - Create SingleChoice Answer", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		wg.Add(1)

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		questionID := random.PositiveInt(t)
		optionID := random.PositiveInt(t)

		// 管理者ユーザーを設定
		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		// 対象ユーザーを設定
		targetUser := &model.User{
			ID:      targetUserID,
			IsStaff: false,
		}

		// 質問 (GetQuestionByIDで返される)
		question := &model.Question{
			Model: gorm.Model{ID: uint(questionID)},
			Title: random.AlphaNumericString(t, 20),
		}

		optionContent := random.AlphaNumericString(t, 20)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), targetUserID).
			Return(targetUser, nil).
			Times(1)

		h.repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(uint(questionID)).
			Return(question, nil).
			Times(1)

		expectedMessage := fmt.Sprintf(
			"@%sがアンケート「%s」のあなたの回答を変更しました。\n### 変更前\n未回答\n### 変更後\n%s",
			adminUserID,
			question.Title,
			optionContent,
		)

		h.traqService.EXPECT().
			PostDirectMessage(gomock.Any(), targetUserID, expectedMessage).
			DoAndReturn(func(_, _, _ any) error {
				defer wg.Done()

				return nil
			}).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			CreateAnswer(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ any, answer *model.Answer) error {
				answer.ID = uint(random.PositiveInt(t))
				// SelectedOptionsにオプションを設定
				answer.SelectedOptions = []model.Option{
					{
						Model:   gorm.Model{ID: uint(optionID)},
						Content: optionContent,
					},
				}

				return nil
			}).
			Times(1)

		reqBody := api.SingleChoiceAnswerRequest{
			Type:       api.SingleChoiceAnswerRequestTypeSingle,
			QuestionId: questionID,
			OptionId:   optionID,
		}

		var req api.AnswerRequest
		err := req.FromSingleChoiceAnswerRequest(reqBody)
		require.NoError(t, err)

		res := h.expect.POST("/api/admin/users/{userId}/answers", targetUserID).
			WithHeader("X-Forwarded-User", adminUserID).
			WithJSON(req).
			Expect().
			Status(http.StatusCreated).JSON().Object()

		waitWithTimeout(t, &wg, 2*time.Second)

		res.Keys().ContainsOnly("id", "type", "userId", "questionId", "selectedOption")
		res.Value("type").String().IsEqual(string(api.SingleChoiceAnswerRequestTypeSingle))
		res.Value("userId").String().IsEqual(targetUserID)
		res.Value("questionId").Number().IsEqual(questionID)
		res.Value("selectedOption").Object().Value("id").Number().IsEqual(optionID)
	})

	t.Run("Success - Create MultipleChoice Answer", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup

		wg.Add(1)

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		questionID := random.PositiveInt(t)
		optionIDs := []int{random.PositiveInt(t), random.PositiveInt(t)}

		// 管理者ユーザーを設定
		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		// 対象ユーザーを設定
		targetUser := &model.User{
			ID:      targetUserID,
			IsStaff: false,
		}

		// 質問 (GetQuestionByIDで返される)
		question := &model.Question{
			Model: gorm.Model{ID: uint(questionID)},
			Title: random.AlphaNumericString(t, 20),
		}

		option1Content := random.AlphaNumericString(t, 20)
		option2Content := random.AlphaNumericString(t, 20)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), targetUserID).
			Return(targetUser, nil).
			Times(1)

		h.repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(uint(questionID)).
			Return(question, nil).
			Times(1)

		expectedMessage := fmt.Sprintf(
			"@%sがアンケート「%s」のあなたの回答を変更しました。\n### 変更前\n未回答\n### 変更後\n- %s\n- %s\n",
			adminUserID,
			question.Title,
			option1Content,
			option2Content,
		)

		h.traqService.EXPECT().
			PostDirectMessage(gomock.Any(), targetUserID, expectedMessage).
			DoAndReturn(func(_, _, _ any) error {
				defer wg.Done()

				return nil
			}).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			CreateAnswer(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ any, answer *model.Answer) error {
				answer.ID = uint(random.PositiveInt(t))
				// SelectedOptionsにオプションを設定
				answer.SelectedOptions = []model.Option{
					{
						Model:   gorm.Model{ID: uint(optionIDs[0])},
						Content: option1Content,
					},
					{
						Model:   gorm.Model{ID: uint(optionIDs[1])},
						Content: option2Content,
					},
				}

				return nil
			}).
			Times(1)

		reqBody := api.MultipleChoiceAnswerRequest{
			Type:       api.MultipleChoiceAnswerRequestTypeMultiple,
			QuestionId: questionID,
			OptionIds:  optionIDs,
		}

		var req api.AnswerRequest
		err := req.FromMultipleChoiceAnswerRequest(reqBody)
		require.NoError(t, err)

		res := h.expect.POST("/api/admin/users/{userId}/answers", targetUserID).
			WithHeader("X-Forwarded-User", adminUserID).
			WithJSON(req).
			Expect().
			Status(http.StatusCreated).JSON().Object()

		waitWithTimeout(t, &wg, 2*time.Second)

		res.Keys().ContainsOnly("id", "type", "userId", "questionId", "selectedOptions")
		res.Value("type").String().IsEqual(string(api.MultipleChoiceAnswerRequestTypeMultiple))
		res.Value("userId").String().IsEqual(targetUserID)
		res.Value("questionId").Number().IsEqual(questionID)
		res.Value("selectedOptions").Array().Length().IsEqual(len(optionIDs))
	})

	t.Run("Forbidden - Non-staff user", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		nonStaffUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		questionID := random.PositiveInt(t)

		// 非スタッフユーザーを設定
		nonStaffUser := &model.User{
			ID:      nonStaffUserID,
			IsStaff: false,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), nonStaffUserID).
			Return(nonStaffUser, nil).
			Times(1)

		reqBody := api.FreeTextAnswerRequest{
			Type:       api.FreeTextAnswerRequestTypeFreeText,
			QuestionId: questionID,
			Content:    random.AlphaNumericString(t, 50),
		}

		var req api.AnswerRequest
		err := req.FromFreeTextAnswerRequest(reqBody)
		require.NoError(t, err)

		h.expect.POST("/api/admin/users/{userId}/answers", targetUserID).
			WithHeader("X-Forwarded-User", nonStaffUserID).
			WithJSON(req).
			Expect().
			Status(http.StatusForbidden).JSON().Object().
			Value("message").String().IsEqual("Forbidden")
	})

	t.Run("BadRequest - Missing X-Forwarded-User header", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		targetUserID := random.AlphaNumericString(t, 32)
		questionID := random.PositiveInt(t)

		reqBody := api.FreeTextAnswerRequest{
			Type:       api.FreeTextAnswerRequestTypeFreeText,
			QuestionId: questionID,
			Content:    random.AlphaNumericString(t, 50),
		}

		var req api.AnswerRequest
		err := req.FromFreeTextAnswerRequest(reqBody)
		require.NoError(t, err)

		h.expect.POST("/api/admin/users/{userId}/answers", targetUserID).
			WithJSON(req).
			Expect().
			Status(http.StatusBadRequest).JSON().Object().
			Value("message").String().IsEqual("X-Forwarded-User header is required")
	})

	t.Run("InternalServerError - Admin user repository error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		questionID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(nil, errors.New("database error")).
			Times(1)

		reqBody := api.FreeTextAnswerRequest{
			Type:       api.FreeTextAnswerRequestTypeFreeText,
			QuestionId: questionID,
			Content:    random.AlphaNumericString(t, 50),
		}

		var req api.AnswerRequest
		err := req.FromFreeTextAnswerRequest(reqBody)
		require.NoError(t, err)

		h.expect.POST("/api/admin/users/{userId}/answers", targetUserID).
			WithHeader("X-Forwarded-User", adminUserID).
			WithJSON(req).
			Expect().
			Status(http.StatusInternalServerError).JSON().Object().
			Value("message").String().IsEqual("Internal server error")
	})

	t.Run("InternalServerError - Target user repository error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		questionID := random.PositiveInt(t)

		// 管理者ユーザーを設定
		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), targetUserID).
			Return(nil, errors.New("database error")).
			Times(1)

		reqBody := api.FreeTextAnswerRequest{
			Type:       api.FreeTextAnswerRequestTypeFreeText,
			QuestionId: questionID,
			Content:    random.AlphaNumericString(t, 50),
		}

		var req api.AnswerRequest
		err := req.FromFreeTextAnswerRequest(reqBody)
		require.NoError(t, err)

		h.expect.POST("/api/admin/users/{userId}/answers", targetUserID).
			WithHeader("X-Forwarded-User", adminUserID).
			WithJSON(req).
			Expect().
			Status(http.StatusInternalServerError).JSON().Object().
			Value("message").String().IsEqual("Internal server error")
	})

	t.Run("InternalServerError - CreateAnswer repository error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		questionID := random.PositiveInt(t)

		// 管理者ユーザーを設定
		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		// 対象ユーザーを設定
		targetUser := &model.User{
			ID:      targetUserID,
			IsStaff: false,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), targetUserID).
			Return(targetUser, nil).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			CreateAnswer(gomock.Any(), gomock.Any()).
			Return(errors.New("database error")).
			Times(1)

		reqBody := api.FreeTextAnswerRequest{
			Type:       api.FreeTextAnswerRequestTypeFreeText,
			QuestionId: questionID,
			Content:    random.AlphaNumericString(t, 50),
		}

		var req api.AnswerRequest
		err := req.FromFreeTextAnswerRequest(reqBody)
		require.NoError(t, err)

		h.expect.POST("/api/admin/users/{userId}/answers", targetUserID).
			WithHeader("X-Forwarded-User", adminUserID).
			WithJSON(req).
			Expect().
			Status(http.StatusInternalServerError).JSON().Object().
			Value("message").String().IsEqual("Internal server error")
	})

	t.Run("NotFound - Question or option not found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		questionID := random.PositiveInt(t)

		// 管理者ユーザーを設定
		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		// 対象ユーザーを設定
		targetUser := &model.User{
			ID:      targetUserID,
			IsStaff: false,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), targetUserID).
			Return(targetUser, nil).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			CreateAnswer(gomock.Any(), gomock.Any()).
			Return(model.ErrNotFound).
			Times(1)

		reqBody := api.FreeTextAnswerRequest{
			Type:       api.FreeTextAnswerRequestTypeFreeText,
			QuestionId: questionID,
			Content:    random.AlphaNumericString(t, 50),
		}

		var req api.AnswerRequest
		err := req.FromFreeTextAnswerRequest(reqBody)
		require.NoError(t, err)

		h.expect.POST("/api/admin/users/{userId}/answers", targetUserID).
			WithHeader("X-Forwarded-User", adminUserID).
			WithJSON(req).
			Expect().
			Status(http.StatusNotFound).JSON().Object().
			Value("message").String().IsEqual("Question or option not found")
	})
}

func TestGetAnswers(t *testing.T) {
	t.Parallel()

	t.Run("Success - Public Question", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		questionID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		freeTextContent := random.AlphaNumericString(t, 50)
		freeNumberContent := random.Float64(t)
		optionID1 := random.PositiveInt(t)
		optionID2 := random.PositiveInt(t)

		answers := []model.Answer{
			{
				Model: gorm.Model{
					ID: uint(random.PositiveInt(t)),
				},
				QuestionID:      questionID,
				UserID:          userID,
				Type:            model.FreeTextQuestion,
				FreeTextContent: &freeTextContent,
			},
			{
				Model: gorm.Model{
					ID: uint(random.PositiveInt(t)),
				},
				QuestionID:        questionID,
				UserID:            userID,
				Type:              model.FreeNumberQuestion,
				FreeNumberContent: &freeNumberContent,
			},
			{
				Model: gorm.Model{
					ID: uint(random.PositiveInt(t)),
				},
				QuestionID: questionID,
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
				QuestionID: questionID,
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

		h.repo.MockAnswerRepository.EXPECT().
			GetAnswers(gomock.Any(), repository.GetAnswersQuery{
				QuestionID:            &questionID,
				IncludePrivateAnswers: false,
			}).
			Return(answers, nil).
			Times(1)

		res := h.expect.GET("/api/questions/{questionId}/answers", questionID).
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

	t.Run("Forbidden - Private Question", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		questionID := uint(random.PositiveInt(t))

		// Private質問の場合はErrForbiddenが返される
		h.repo.MockAnswerRepository.EXPECT().
			GetAnswers(gomock.Any(), repository.GetAnswersQuery{
				QuestionID:            &questionID,
				IncludePrivateAnswers: false,
			}).
			Return(nil, model.ErrForbidden).
			Times(1)

		h.expect.GET("/api/questions/{questionId}/answers", questionID).
			Expect().
			Status(http.StatusForbidden).JSON().Object().
			Value("message").String().IsEqual("Question is not public")
	})

	t.Run("Success - Public Question with No Answers", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		questionID := uint(random.PositiveInt(t))

		// 回答が存在しない場合（Public質問だが回答がない）
		h.repo.MockAnswerRepository.EXPECT().
			GetAnswers(gomock.Any(), repository.GetAnswersQuery{
				QuestionID:            &questionID,
				IncludePrivateAnswers: false,
			}).
			Return([]model.Answer{}, nil).
			Times(1)

		res := h.expect.GET("/api/questions/{questionId}/answers", questionID).
			Expect().
			Status(http.StatusOK).JSON().Array()

		res.Length().IsEqual(0)
	})

	t.Run("Not Found - Question Does Not Exist", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		questionID := uint(random.PositiveInt(t))

		// 質問が存在しない場合はErrNotFoundが返される
		h.repo.MockAnswerRepository.EXPECT().
			GetAnswers(gomock.Any(), repository.GetAnswersQuery{
				QuestionID:            &questionID,
				IncludePrivateAnswers: false,
			}).
			Return(nil, model.ErrNotFound).
			Times(1)

		h.expect.GET("/api/questions/{questionId}/answers", questionID).
			Expect().
			Status(http.StatusNotFound).JSON().Object().
			Value("message").String().IsEqual("Question not found")
	})

	t.Run("Internal Server Error - Repository Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		questionID := uint(random.PositiveInt(t))

		h.repo.MockAnswerRepository.EXPECT().
			GetAnswers(gomock.Any(), repository.GetAnswersQuery{
				QuestionID:            &questionID,
				IncludePrivateAnswers: false,
			}).
			Return(nil, errors.New("repository error")).
			Times(1)

		h.expect.GET("/api/questions/{questionId}/answers", questionID).
			Expect().
			Status(http.StatusInternalServerError).JSON().Object().
			Value("message").String().IsEqual("Internal server error")
	})
}

func TestAdminGetAnswersForQuestionGroup(t *testing.T) {
	t.Parallel()

	t.Run("Success with specific user", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		questionGroupID := uint(random.PositiveInt(t))

		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		answers := []model.Answer{
			{
				Model: gorm.Model{
					ID: uint(random.PositiveInt(t)),
				},
				UserID:          targetUserID,
				QuestionID:      uint(random.PositiveInt(t)),
				Type:            model.FreeTextQuestion,
				FreeTextContent: &[]string{random.AlphaNumericString(t, 100)}[0],
			},
			{
				Model: gorm.Model{
					ID: uint(random.PositiveInt(t)),
				},
				UserID:          targetUserID,
				QuestionID:      uint(random.PositiveInt(t)),
				Type:            model.FreeTextQuestion,
				FreeTextContent: &[]string{random.AlphaNumericString(t, 100)}[0],
			},
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			GetAnswers(gomock.Any(), repository.GetAnswersQuery{
				UserID:                &targetUserID,
				QuestionGroupID:       &questionGroupID,
				IncludePrivateAnswers: true,
			}).
			Return(answers, nil).
			Times(1)

		h.expect.GET("/api/admin/question-groups/{questionGroupId}/answers", questionGroupID).
			WithQuery("userId", targetUserID).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().
			Status(http.StatusOK).JSON().Array().Length().IsEqual(2)
	})

	t.Run("Success with all users", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		questionGroupID := uint(random.PositiveInt(t))

		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		answers := []model.Answer{
			{
				Model: gorm.Model{
					ID: uint(random.PositiveInt(t)),
				},
				UserID:          random.AlphaNumericString(t, 32),
				QuestionID:      uint(random.PositiveInt(t)),
				Type:            model.FreeTextQuestion,
				FreeTextContent: &[]string{random.AlphaNumericString(t, 100)}[0],
			},
			{
				Model: gorm.Model{
					ID: uint(random.PositiveInt(t)),
				},
				UserID:          random.AlphaNumericString(t, 32),
				QuestionID:      uint(random.PositiveInt(t)),
				Type:            model.FreeTextQuestion,
				FreeTextContent: &[]string{random.AlphaNumericString(t, 100)}[0],
			},
			{
				Model: gorm.Model{
					ID: uint(random.PositiveInt(t)),
				},
				UserID:          random.AlphaNumericString(t, 32),
				QuestionID:      uint(random.PositiveInt(t)),
				Type:            model.FreeTextQuestion,
				FreeTextContent: &[]string{random.AlphaNumericString(t, 100)}[0],
			},
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			GetAnswers(gomock.Any(), repository.GetAnswersQuery{
				QuestionGroupID:       &questionGroupID,
				IncludePrivateAnswers: true,
			}).
			Return(answers, nil).
			Times(1)

		h.expect.GET("/api/admin/question-groups/{questionGroupId}/answers", questionGroupID).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().
			Status(http.StatusOK).JSON().Array().Length().IsEqual(3)
	})

	t.Run("Forbidden - User is not staff", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		questionGroupID := random.PositiveInt(t)

		user := &model.User{
			ID:      userID,
			IsStaff: false,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(user, nil).
			Times(1)

		h.expect.GET("/api/admin/question-groups/{questionGroupId}/answers", questionGroupID).
			WithQuery("userId", targetUserID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusForbidden).JSON().Object().
			Value("message").String().IsEqual("Forbidden")
	})

	t.Run("Bad Request - Missing X-Forwarded-User header", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		targetUserID := random.AlphaNumericString(t, 32)
		questionGroupID := random.PositiveInt(t)

		h.expect.GET("/api/admin/question-groups/{questionGroupId}/answers", questionGroupID).
			WithQuery("userId", targetUserID).
			Expect().
			Status(http.StatusBadRequest).JSON().Object().
			Value("message").String().IsEqual("X-Forwarded-User header is required")
	})

	t.Run("Internal Server Error - GetOrCreateUser Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		questionGroupID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(nil, errors.New("repository error")).
			Times(1)

		h.expect.GET("/api/admin/question-groups/{questionGroupId}/answers", questionGroupID).
			WithQuery("userId", targetUserID).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().
			Status(http.StatusInternalServerError).JSON().Object().
			Value("message").String().IsEqual("Internal server error")
	})

	t.Run("Internal Server Error - GetAnswersByUserAndQuestionGroup Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		questionGroupID := uint(random.PositiveInt(t))

		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			GetAnswers(gomock.Any(), repository.GetAnswersQuery{
				UserID:                &targetUserID,
				QuestionGroupID:       &questionGroupID,
				IncludePrivateAnswers: true,
			}).
			Return(nil, errors.New("repository error")).
			Times(1)

		h.expect.GET("/api/admin/question-groups/{questionGroupId}/answers", questionGroupID).
			WithQuery("userId", targetUserID).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().
			Status(http.StatusInternalServerError).JSON().Object().
			Value("message").String().IsEqual("Internal server error")
	})

	t.Run("Internal Server Error - GetAnswersByQuestionGroup Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		questionGroupID := uint(random.PositiveInt(t))

		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			GetAnswers(gomock.Any(), repository.GetAnswersQuery{
				QuestionGroupID:       &questionGroupID,
				IncludePrivateAnswers: true,
			}).
			Return(nil, errors.New("repository error")).
			Times(1)

		h.expect.GET("/api/admin/question-groups/{questionGroupId}/answers", questionGroupID).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().
			Status(http.StatusInternalServerError).JSON().Object().
			Value("message").String().IsEqual("Internal server error")
	})

	t.Run("Not Found - Question group does not exist", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		questionGroupID := uint(random.PositiveInt(t))

		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			GetAnswers(gomock.Any(), repository.GetAnswersQuery{
				QuestionGroupID:       &questionGroupID,
				IncludePrivateAnswers: true,
			}).
			Return(nil, model.ErrNotFound).
			Times(1)

		h.expect.GET("/api/admin/question-groups/{questionGroupId}/answers", questionGroupID).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().
			Status(http.StatusNotFound).JSON().Object().
			Value("message").String().IsEqual("Question group not found")
	})

	t.Run("Not Found - Question group does not exist with specific user", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		questionGroupID := uint(random.PositiveInt(t))

		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)

		h.repo.MockAnswerRepository.EXPECT().
			GetAnswers(gomock.Any(), repository.GetAnswersQuery{
				UserID:                &targetUserID,
				QuestionGroupID:       &questionGroupID,
				IncludePrivateAnswers: true,
			}).
			Return(nil, model.ErrNotFound).
			Times(1)

		h.expect.GET("/api/admin/question-groups/{questionGroupId}/answers", questionGroupID).
			WithQuery("userId", targetUserID).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().
			Status(http.StatusNotFound).JSON().Object().
			Value("message").String().IsEqual("Question group not found")
	})
}

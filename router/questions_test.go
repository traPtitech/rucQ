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

func TestAdminPostQuestion(t *testing.T) {
	t.Parallel()

	t.Run("Success (Single Choice)", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		singleChoiceQuestion := api.PostSingleChoiceQuestionRequest{
			Type:        api.PostSingleChoiceQuestionRequestTypeSingle,
			Title:       random.AlphaNumericString(t, 10),
			Description: random.PtrOrNil(t, random.AlphaNumericString(t, 20)),
			IsPublic:    random.Bool(t),
			IsOpen:      random.Bool(t),
			IsRequired:  random.PtrOrNil(t, random.Bool(t)),
			Options: []api.PostOptionRequest{
				{
					Content: random.AlphaNumericString(t, 10),
				},
			},
		}

		var req api.PostQuestionRequest

		err := req.FromPostSingleChoiceQuestionRequest(singleChoiceQuestion)

		require.NoError(t, err)

		userID := random.AlphaNumericString(t, 32)
		questionGroupID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().GetOrCreateUser(gomock.Any(), userID).Return(&model.User{
			IsStaff: true,
		}, nil).Times(1)
		h.repo.MockQuestionRepository.EXPECT().
			CreateQuestion(gomock.Any()).
			Return(nil).
			Times(1)

		res := h.expect.POST("/api/admin/question-groups/{questionGroupID}/questions", questionGroupID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object()

		res.Keys().ContainsAll("id", "type", "title", "isPublic", "isOpen", "isRequired", "options")
		res.Value("type").IsEqual(api.PostSingleChoiceQuestionRequestTypeSingle)
		res.Value("title").IsEqual(singleChoiceQuestion.Title)

		if singleChoiceQuestion.Description != nil {
			res.Value("description").IsEqual(*singleChoiceQuestion.Description)
		} else {
			res.Keys().NotContainsAny("description")
		}

		res.Value("isPublic").IsEqual(singleChoiceQuestion.IsPublic)
		res.Value("isOpen").IsEqual(singleChoiceQuestion.IsOpen)
		
		if singleChoiceQuestion.IsRequired != nil {
			res.Value("isRequired").IsEqual(*singleChoiceQuestion.IsRequired)
		} else {
			// デフォルト値はfalseになることを確認
			res.Value("isRequired").IsEqual(false)
		}

		options := res.Value("options").Array()

		options.Length().IsEqual(len(singleChoiceQuestion.Options))

		option := options.Value(0).Object()

		option.Keys().ContainsOnly("id", "content")
		option.Value("content").IsEqual(singleChoiceQuestion.Options[0].Content)
	})
}

func TestAdminPutQuestion(t *testing.T) {
	t.Parallel()

	t.Run("Success (Single Choice)", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		questionID := uint(random.PositiveInt(t))
		title := random.AlphaNumericString(t, 15)
		description := random.PtrOrNil(t, random.AlphaNumericString(t, 25))
		isPublic := random.Bool(t)
		isOpen := random.Bool(t)
		isRequired := random.PtrOrNil(t, random.Bool(t))
		optionContent := random.AlphaNumericString(t, 10)

		req := api.PutSingleChoiceQuestionRequest{
			Type:        api.PutSingleChoiceQuestionRequestTypeSingle,
			Title:       title,
			Description: description,
			IsPublic:    isPublic,
			IsOpen:      isOpen,
			IsRequired:  isRequired,
			Options: []api.PutOptionRequest{
				{
					Content: optionContent,
				},
			},
		}

		userID := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().GetOrCreateUser(gomock.Any(), userID).Return(&model.User{
			IsStaff: true,
		}, nil).Times(1)
		h.repo.MockQuestionRepository.EXPECT().
			UpdateQuestion(gomock.Any(), questionID, gomock.Any()).
			Return(nil).
			Times(1)

		res := h.expect.PUT("/api/admin/questions/{questionID}", questionID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		res.Keys().ContainsAll("id", "type", "title", "isPublic", "isOpen", "isRequired", "options")
		res.Value("type").IsEqual(api.PutSingleChoiceQuestionRequestTypeSingle)
		res.Value("title").IsEqual(title)

		if description != nil {
			res.Value("description").IsEqual(*description)
		} else {
			res.Keys().NotContainsAny("description")
		}

		res.Value("isPublic").IsEqual(isPublic)
		res.Value("isOpen").IsEqual(isOpen)
		
		if isRequired != nil {
			res.Value("isRequired").IsEqual(*isRequired)
		} else {
			res.Value("isRequired").IsEqual(false)
		}

		options := res.Value("options").Array()

		options.Length().IsEqual(len(req.Options))

		option := options.Value(0).Object()

		option.Keys().ContainsOnly("id", "content")
		option.Value("content").IsEqual(optionContent)
	})
}

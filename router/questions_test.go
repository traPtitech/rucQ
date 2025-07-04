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

		res.Keys().ContainsAll("id", "type", "title", "isPublic", "isOpen", "options")
		res.Value("type").IsEqual(api.PostSingleChoiceQuestionRequestTypeSingle)
		res.Value("title").IsEqual(singleChoiceQuestion.Title)

		if singleChoiceQuestion.Description != nil {
			res.Value("description").IsEqual(*singleChoiceQuestion.Description)
		} else {
			res.Keys().NotContainsAny("description")
		}

		res.Value("isPublic").IsEqual(singleChoiceQuestion.IsPublic)
		res.Value("isOpen").IsEqual(singleChoiceQuestion.IsOpen)

		options := res.Value("options").Array()

		options.Length().IsEqual(len(singleChoiceQuestion.Options))

		option := options.Value(0).Object()

		option.Keys().ContainsOnly("id", "content")
		option.Value("content").IsEqual(singleChoiceQuestion.Options[0].Content)
	})
}

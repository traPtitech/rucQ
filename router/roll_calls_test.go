package router

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestServer_GetRollCalls(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))
		user1 := model.User{ID: random.AlphaNumericString(t, 32)}
		user2 := model.User{ID: random.AlphaNumericString(t, 32)}

		rollCall1 := model.RollCall{
			Model: gorm.Model{
				ID: uint(random.PositiveInt(t)),
			},
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Options: []string{
				random.AlphaNumericString(t, 5),
				random.AlphaNumericString(t, 5),
				random.AlphaNumericString(t, 5),
			},
			Subjects: []model.User{user1, user2},
			CampID:   campID,
		}

		rollCall2 := model.RollCall{
			Model: gorm.Model{
				ID: uint(random.PositiveInt(t)),
			},
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Options: []string{
				random.AlphaNumericString(t, 5),
				random.AlphaNumericString(t, 5),
			},
			Subjects: []model.User{user1},
			CampID:   campID,
		}

		h.repo.MockRollCallRepository.EXPECT().
			GetRollCalls(gomock.Any(), campID).
			Return([]model.RollCall{
				rollCall1,
				rollCall2,
			}, nil).
			Times(1)

		res := h.expect.GET("/api/camps/{campId}/roll-calls", campID).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		res.Length().IsEqual(2)

		res1 := res.Value(0).Object()
		res1.Keys().ContainsOnly("id", "name", "description", "options", "subjects")
		res1.Value("id").Number().IsEqual(rollCall1.ID)
		res1.Value("name").String().IsEqual(rollCall1.Name)
		res1.Value("description").String().IsEqual(rollCall1.Description)
		res1.Value("options").Array().IsEqual(rollCall1.Options)
		res1.Value("subjects").Array().IsEqual([]string{user1.ID, user2.ID})

		res2 := res.Value(1).Object()
		res2.Keys().ContainsOnly("id", "name", "description", "options", "subjects")
		res2.Value("id").Number().IsEqual(rollCall2.ID)
		res2.Value("name").String().IsEqual(rollCall2.Name)
		res2.Value("description").String().IsEqual(rollCall2.Description)
		res2.Value("options").Array().IsEqual(rollCall2.Options)
		res2.Value("subjects").Array().IsEqual([]string{user1.ID})
	})

	t.Run("Camp not found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))

		h.repo.MockRollCallRepository.EXPECT().GetRollCalls(gomock.Any(), campID).Return(
			nil, repository.ErrCampNotFound,
		).Times(1)

		h.expect.GET("/api/camps/{campId}/roll-calls", campID).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			Value("message").String().IsEqual("Camp not found")
	})

	t.Run("Repository error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))

		h.repo.MockRollCallRepository.EXPECT().GetRollCalls(gomock.Any(), campID).Return(
			nil, errors.New("repository error"),
		).Times(1)

		h.expect.GET("/api/camps/{campId}/roll-calls", campID).
			Expect().
			Status(http.StatusInternalServerError)
	})
}

func TestServer_AdminPostRollCall(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		user := model.User{
			ID:      userID,
			IsStaff: true,
		}

		requestBody := api.RollCallRequest{
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Options: []string{
				random.AlphaNumericString(t, 5),
				random.AlphaNumericString(t, 5),
				random.AlphaNumericString(t, 5),
			},
			Subjects: []string{
				random.AlphaNumericString(t, 32),
				random.AlphaNumericString(t, 32),
			},
		}

		h.repo.MockUserRepository.EXPECT().GetOrCreateUser(gomock.Any(), userID).Return(&user, nil)
		h.repo.MockRollCallRepository.EXPECT().
			CreateRollCall(gomock.Any(), gomock.Any()).
			Return(nil).
			Do(func(_, rollCall any) {
				rc := rollCall.(*model.RollCall)
				rc.ID = uint(random.PositiveInt(t))
			}).Times(1)

		res := h.expect.POST("/api/admin/camps/{campId}/roll-calls", campID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(requestBody).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object()

		res.Keys().ContainsOnly("id", "name", "description", "options", "subjects")
		res.Value("name").String().IsEqual(requestBody.Name)
		res.Value("description").String().IsEqual(requestBody.Description)
		res.Value("options").Array().IsEqual(requestBody.Options)
		res.Value("subjects").Array().IsEqual(requestBody.Subjects)
	})

	t.Run("Missing X-Forwarded-User header", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))

		requestBody := api.RollCallRequest{
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Options: []string{
				random.AlphaNumericString(t, 5),
				random.AlphaNumericString(t, 5),
			},
			Subjects: []string{
				random.AlphaNumericString(t, 32),
			},
		}

		h.expect.POST("/api/admin/camps/{campId}/roll-calls", campID).
			WithJSON(requestBody).
			Expect().
			Status(http.StatusBadRequest).
			JSON().
			Object().
			Value("message").String().IsEqual("X-Forwarded-User header is required")
	})

	t.Run("User not staff", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		user := model.User{
			ID:      userID,
			IsStaff: false,
		}

		requestBody := api.RollCallRequest{
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Options: []string{
				random.AlphaNumericString(t, 5),
				random.AlphaNumericString(t, 5),
			},
			Subjects: []string{
				random.AlphaNumericString(t, 32),
			},
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&user, nil).
			Times(1)

		h.expect.POST("/api/admin/camps/{campId}/roll-calls", campID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(requestBody).
			Expect().
			Status(http.StatusForbidden).
			JSON().
			Object().
			Value("message").String().IsEqual("Forbidden")
	})

	t.Run("Camp not found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		user := model.User{
			ID:      userID,
			IsStaff: true,
		}

		requestBody := api.RollCallRequest{
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Options: []string{
				random.AlphaNumericString(t, 5),
				random.AlphaNumericString(t, 5),
			},
			Subjects: []string{
				random.AlphaNumericString(t, 32),
			},
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&user, nil).
			Times(1)
		h.repo.MockRollCallRepository.EXPECT().
			CreateRollCall(gomock.Any(), gomock.Any()).
			Return(repository.ErrCampNotFound).
			Times(1)

		h.expect.POST("/api/admin/camps/{campId}/roll-calls", campID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(requestBody).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			Value("message").String().IsEqual("Camp not found")
	})

	t.Run("Subject user not found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		user := model.User{
			ID:      userID,
			IsStaff: true,
		}

		requestBody := api.RollCallRequest{
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Options: []string{
				random.AlphaNumericString(t, 5),
				random.AlphaNumericString(t, 5),
			},
			Subjects: []string{
				random.AlphaNumericString(t, 32),
			},
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&user, nil).
			Times(1)

		h.repo.MockRollCallRepository.EXPECT().
			CreateRollCall(gomock.Any(), gomock.Any()).
			Return(repository.ErrUserNotFound).
			Times(1)

		h.expect.POST("/api/admin/camps/{campId}/roll-calls", campID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(requestBody).
			Expect().
			Status(http.StatusBadRequest).
			JSON().
			Object().
			Value("message").String().IsEqual("One or more subject users not found")
	})

	t.Run("GetOrCreateUser Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		requestBody := api.RollCallRequest{
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Options: []string{
				random.AlphaNumericString(t, 5),
				random.AlphaNumericString(t, 5),
			},
			Subjects: []string{
				random.AlphaNumericString(t, 32),
			},
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(nil, errors.New("user repository error")).
			Times(1)

		h.expect.POST("/api/admin/camps/{campId}/roll-calls", campID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(requestBody).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("CreateRollCall repository error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		user := model.User{
			ID:      userID,
			IsStaff: true,
		}

		requestBody := api.RollCallRequest{
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Options: []string{
				random.AlphaNumericString(t, 5),
				random.AlphaNumericString(t, 5),
			},
			Subjects: []string{
				random.AlphaNumericString(t, 32),
			},
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&user, nil).
			Times(1)

		h.repo.MockRollCallRepository.EXPECT().
			CreateRollCall(gomock.Any(), gomock.Any()).
			Return(errors.New("create roll call error")).
			Times(1)

		h.expect.POST("/api/admin/camps/{campId}/roll-calls", campID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(requestBody).
			Expect().
			Status(http.StatusInternalServerError)
	})
}

func TestServer_GetRollCallReactions(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		rollCallID := uint(random.PositiveInt(t))
		user1 := random.AlphaNumericString(t, 32)
		user2 := random.AlphaNumericString(t, 32)

		reaction1 := model.RollCallReaction{
			Model:      gorm.Model{ID: uint(random.PositiveInt(t))},
			Content:    random.AlphaNumericString(t, 10),
			UserID:     user1,
			RollCallID: rollCallID,
		}

		reaction2 := model.RollCallReaction{
			Model:      gorm.Model{ID: uint(random.PositiveInt(t))},
			Content:    random.AlphaNumericString(t, 15),
			UserID:     user2,
			RollCallID: rollCallID,
		}

		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactions(gomock.Any(), rollCallID).
			Return([]model.RollCallReaction{reaction1, reaction2}, nil).
			Times(1)

		res := h.expect.GET("/api/roll-calls/{rollCallId}/reactions", rollCallID).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		res.Length().IsEqual(2)

		res1 := res.Value(0).Object()
		res1.Keys().ContainsOnly("id", "content", "userId")
		res1.Value("id").Number().IsEqual(reaction1.ID)
		res1.Value("content").String().IsEqual(reaction1.Content)
		res1.Value("userId").String().IsEqual(reaction1.UserID)

		res2 := res.Value(1).Object()
		res2.Keys().ContainsOnly("id", "content", "userId")
		res2.Value("id").Number().IsEqual(reaction2.ID)
		res2.Value("content").String().IsEqual(reaction2.Content)
		res2.Value("userId").String().IsEqual(reaction2.UserID)
	})

	t.Run("Roll call not found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		rollCallID := uint(random.PositiveInt(t))

		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactions(gomock.Any(), rollCallID).
			Return(nil, repository.ErrRollCallNotFound).
			Times(1)

		h.expect.GET("/api/roll-calls/{rollCallId}/reactions", rollCallID).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			Value("message").String().IsEqual("Roll call not found")
	})

	t.Run("Repository error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		rollCallID := uint(random.PositiveInt(t))

		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactions(gomock.Any(), rollCallID).
			Return(nil, errors.New("repository error")).
			Times(1)

		h.expect.GET("/api/roll-calls/{rollCallId}/reactions", rollCallID).
			Expect().
			Status(http.StatusInternalServerError)
	})
}

func TestServer_PostRollCallReaction(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		rollCallID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		user := model.User{ID: userID}

		requestBody := api.PostRollCallReactionJSONRequestBody{
			Content: random.AlphaNumericString(t, 20),
		}

		expectedReaction := model.RollCallReaction{
			Model:      gorm.Model{ID: uint(random.PositiveInt(t))},
			Content:    requestBody.Content,
			UserID:     userID,
			RollCallID: rollCallID,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&user, nil).
			Times(1)

		h.repo.MockRollCallReactionRepository.EXPECT().
			CreateRollCallReaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ any, reaction *model.RollCallReaction) error {
				*reaction = expectedReaction
				return nil
			}).
			Times(1)

		res := h.expect.POST("/api/roll-calls/{rollCallId}/reactions", rollCallID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(requestBody).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object()

		res.Keys().ContainsOnly("id", "content", "userId")
		res.Value("id").Number().IsEqual(expectedReaction.ID)
		res.Value("content").String().IsEqual(expectedReaction.Content)
		res.Value("userId").String().IsEqual(expectedReaction.UserID)
	})

	t.Run("Missing X-Forwarded-User header", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		rollCallID := uint(random.PositiveInt(t))

		requestBody := api.PostRollCallReactionJSONRequestBody{
			Content: random.AlphaNumericString(t, 20),
		}

		h.expect.POST("/api/roll-calls/{rollCallId}/reactions", rollCallID).
			WithJSON(requestBody).
			Expect().
			Status(http.StatusBadRequest).
			JSON().
			Object().
			Value("message").String().IsEqual("X-Forwarded-User header is required")
	})

	t.Run("Roll call not found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		rollCallID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		user := model.User{ID: userID}

		requestBody := api.PostRollCallReactionJSONRequestBody{
			Content: random.AlphaNumericString(t, 20),
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&user, nil).
			Times(1)

		h.repo.MockRollCallReactionRepository.EXPECT().
			CreateRollCallReaction(gomock.Any(), gomock.Any()).
			Return(repository.ErrRollCallNotFound).
			Times(1)

		h.expect.POST("/api/roll-calls/{rollCallId}/reactions", rollCallID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(requestBody).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			Value("message").String().IsEqual("Roll call not found")
	})

	t.Run("GetOrCreateUser Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		rollCallID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		requestBody := api.PostRollCallReactionJSONRequestBody{
			Content: random.AlphaNumericString(t, 20),
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(nil, errors.New("user repository error")).
			Times(1)

		h.expect.POST("/api/roll-calls/{rollCallId}/reactions", rollCallID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(requestBody).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("CreateRollCallReaction repository error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		rollCallID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		user := model.User{ID: userID}

		requestBody := api.PostRollCallReactionJSONRequestBody{
			Content: random.AlphaNumericString(t, 20),
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&user, nil).
			Times(1)

		h.repo.MockRollCallReactionRepository.EXPECT().
			CreateRollCallReaction(gomock.Any(), gomock.Any()).
			Return(errors.New("create reaction error")).
			Times(1)

		h.expect.POST("/api/roll-calls/{rollCallId}/reactions", rollCallID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(requestBody).
			Expect().
			Status(http.StatusInternalServerError)
	})
}

func TestServer_PutReaction(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		reactionID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		existingReaction := model.RollCallReaction{
			Model:      gorm.Model{ID: reactionID},
			Content:    random.AlphaNumericString(t, 10),
			UserID:     userID,
			RollCallID: uint(random.PositiveInt(t)),
		}

		requestBody := api.PutReactionJSONRequestBody{
			Content: random.AlphaNumericString(t, 20),
		}

		updatedReaction := existingReaction
		updatedReaction.Content = requestBody.Content

		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactionByID(gomock.Any(), reactionID).
			Return(&existingReaction, nil).
			Times(1) // 最初の確認のみ

		h.repo.MockRollCallReactionRepository.EXPECT().
			UpdateRollCallReaction(gomock.Any(), reactionID, gomock.Any()).
			Return(nil).
			Times(1)

		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactionByID(gomock.Any(), reactionID).
			Return(&updatedReaction, nil).
			Times(1) // 更新後の取得

		res := h.expect.PUT("/api/reactions/{reactionId}", reactionID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(requestBody).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		res.Keys().ContainsOnly("id", "content", "userId")
		res.Value("id").Number().IsEqual(reactionID)
		res.Value("content").String().IsEqual(requestBody.Content)
		res.Value("userId").String().IsEqual(userID)
	})

	t.Run("Reaction not found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		reactionID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		requestBody := api.PutReactionJSONRequestBody{
			Content: random.AlphaNumericString(t, 20),
		}

		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactionByID(gomock.Any(), reactionID).
			Return(nil, repository.ErrRollCallReactionNotFound).
			Times(1)

		h.expect.PUT("/api/reactions/{reactionId}", reactionID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(requestBody).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			Value("message").String().IsEqual("Reaction not found")
	})

	t.Run("Forbidden - not owner", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		reactionID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		otherUserID := random.AlphaNumericString(t, 32)

		existingReaction := model.RollCallReaction{
			Model:      gorm.Model{ID: reactionID},
			Content:    random.AlphaNumericString(t, 10),
			UserID:     otherUserID, // 異なるユーザー
			RollCallID: uint(random.PositiveInt(t)),
		}

		requestBody := api.PutReactionJSONRequestBody{
			Content: random.AlphaNumericString(t, 20),
		}

		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactionByID(gomock.Any(), reactionID).
			Return(&existingReaction, nil).
			Times(1)

		h.expect.PUT("/api/reactions/{reactionId}", reactionID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(requestBody).
			Expect().
			Status(http.StatusForbidden).
			JSON().
			Object().
			Value("message").String().IsEqual("You can only edit your own reactions")
	})

	t.Run("GetRollCallReactionByID error - initial check", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		reactionID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		requestBody := api.PutReactionJSONRequestBody{
			Content: random.AlphaNumericString(t, 20),
		}

		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactionByID(gomock.Any(), reactionID).
			Return(nil, errors.New("get reaction error")).
			Times(1)

		h.expect.PUT("/api/reactions/{reactionId}", reactionID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(requestBody).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("UpdateRollCallReaction error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		reactionID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		existingReaction := model.RollCallReaction{
			Model:      gorm.Model{ID: reactionID},
			Content:    random.AlphaNumericString(t, 10),
			UserID:     userID,
			RollCallID: uint(random.PositiveInt(t)),
		}

		requestBody := api.PutReactionJSONRequestBody{
			Content: random.AlphaNumericString(t, 20),
		}

		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactionByID(gomock.Any(), reactionID).
			Return(&existingReaction, nil).
			Times(1)

		h.repo.MockRollCallReactionRepository.EXPECT().
			UpdateRollCallReaction(gomock.Any(), reactionID, gomock.Any()).
			Return(errors.New("update reaction error")).
			Times(1)

		h.expect.PUT("/api/reactions/{reactionId}", reactionID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(requestBody).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("GetRollCallReactionByID error - after update", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		reactionID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		existingReaction := model.RollCallReaction{
			Model:      gorm.Model{ID: reactionID},
			Content:    random.AlphaNumericString(t, 10),
			UserID:     userID,
			RollCallID: uint(random.PositiveInt(t)),
		}

		requestBody := api.PutReactionJSONRequestBody{
			Content: random.AlphaNumericString(t, 20),
		}

		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactionByID(gomock.Any(), reactionID).
			Return(&existingReaction, nil).
			Times(1)

		h.repo.MockRollCallReactionRepository.EXPECT().
			UpdateRollCallReaction(gomock.Any(), reactionID, gomock.Any()).
			Return(nil).
			Times(1)

		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactionByID(gomock.Any(), reactionID).
			Return(nil, errors.New("get updated reaction error")).
			Times(1)

		h.expect.PUT("/api/reactions/{reactionId}", reactionID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(requestBody).
			Expect().
			Status(http.StatusInternalServerError)
	})
}

func TestServer_DeleteReaction(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		reactionID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		existingReaction := model.RollCallReaction{
			Model:      gorm.Model{ID: reactionID},
			Content:    random.AlphaNumericString(t, 10),
			UserID:     userID,
			RollCallID: uint(random.PositiveInt(t)),
		}

		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactionByID(gomock.Any(), reactionID).
			Return(&existingReaction, nil).
			Times(1)

		h.repo.MockRollCallReactionRepository.EXPECT().
			DeleteRollCallReaction(gomock.Any(), reactionID).
			Return(nil).
			Times(1)

		h.expect.DELETE("/api/reactions/{reactionId}", reactionID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusNoContent).
			NoContent()
	})

	t.Run("Reaction not found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		reactionID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactionByID(gomock.Any(), reactionID).
			Return(nil, repository.ErrRollCallReactionNotFound).
			Times(1)

		h.expect.DELETE("/api/reactions/{reactionId}", reactionID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			Value("message").String().IsEqual("Reaction not found")
	})

	t.Run("Forbidden - not owner", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		reactionID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		otherUserID := random.AlphaNumericString(t, 32)

		existingReaction := model.RollCallReaction{
			Model:      gorm.Model{ID: reactionID},
			Content:    random.AlphaNumericString(t, 10),
			UserID:     otherUserID, // 異なるユーザー
			RollCallID: uint(random.PositiveInt(t)),
		}

		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactionByID(gomock.Any(), reactionID).
			Return(&existingReaction, nil).
			Times(1)

		h.expect.DELETE("/api/reactions/{reactionId}", reactionID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusForbidden).
			JSON().
			Object().
			Value("message").String().IsEqual("You can only delete your own reactions")
	})

	t.Run("GetRollCallReactionByID error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		reactionID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactionByID(gomock.Any(), reactionID).
			Return(nil, errors.New("get reaction error")).
			Times(1)

		h.expect.DELETE("/api/reactions/{reactionId}", reactionID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("DeleteRollCallReaction error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		reactionID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		existingReaction := model.RollCallReaction{
			Model:      gorm.Model{ID: reactionID},
			Content:    random.AlphaNumericString(t, 10),
			UserID:     userID,
			RollCallID: uint(random.PositiveInt(t)),
		}

		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactionByID(gomock.Any(), reactionID).
			Return(&existingReaction, nil).
			Times(1)

		h.repo.MockRollCallReactionRepository.EXPECT().
			DeleteRollCallReaction(gomock.Any(), reactionID).
			Return(errors.New("delete reaction error")).
			Times(1)

		h.expect.DELETE("/api/reactions/{reactionId}", reactionID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusInternalServerError)
	})
}

const eventStreamDataPrefix = "data: "

func TestServer_StreamRollCallReactions(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		rollCallID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		user := model.User{ID: userID}

		// SSEリクエストの準備
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)

		defer cancel()

		req := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			fmt.Sprintf("/api/roll-calls/%d/reactions/stream", rollCallID),
			nil,
		)
		rec := httptest.NewRecorder()

		// SSEハンドラを非同期で実行
		go func() {
			h.e.ServeHTTP(rec, req)
		}()

		// レスポンスヘッダーの確認
		assert.Eventually(t, func() bool {
			return rec.Header().Get(echo.HeaderContentType) == "text/event-stream"
		}, 2*time.Second, 50*time.Millisecond, "content-type not text/event-stream", rec.Header().Get(echo.HeaderContentType))

		// イベントを読み取るためのスキャナ
		scanner := bufio.NewScanner(rec.Body)

		// 1. Create Reaction
		originalContent := random.AlphaNumericString(t, 20)
		createReqBody := api.PostRollCallReactionJSONRequestBody{Content: originalContent}
		reactionID := uint(random.PositiveInt(t))
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&user, nil).
			Times(1)
		h.repo.MockRollCallReactionRepository.EXPECT().
			CreateRollCallReaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, reaction *model.RollCallReaction) error {
				reaction.ID = reactionID

				return nil
			}).
			Times(1)

		h.expect.POST("/api/roll-calls/{rollCallId}/reactions", rollCallID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(createReqBody).
			Expect().
			Status(http.StatusCreated)

		// SSEイベントの受信と検証 (Created)
		if assert.True(t, scanner.Scan()) {
			line := scanner.Text()

			if assert.True(
				t,
				strings.HasPrefix(line, eventStreamDataPrefix),
				"line not start with 'data: '",
				line,
			) {
				assert.JSONEq(
					t,
					fmt.Sprintf(
						`{"id":%d,"type":"created","userId":"%s","content":"%s"}`,
						reactionID,
						userID,
						originalContent,
					),
					strings.TrimPrefix(line, eventStreamDataPrefix),
				)
			}
		}

		if assert.True(t, scanner.Scan()) {
			assert.Empty(t, scanner.Text())
		}

		// 2. Update Reaction
		updatedContent := random.AlphaNumericString(t, 20)
		updateReqBody := api.PutReactionJSONRequestBody{Content: updatedContent}
		existingReaction := model.RollCallReaction{
			Model:      gorm.Model{ID: reactionID},
			UserID:     userID,
			RollCallID: rollCallID,
			Content:    originalContent,
		}
		updatedReaction := model.RollCallReaction{
			Model:      gorm.Model{ID: reactionID},
			UserID:     userID,
			RollCallID: rollCallID,
			Content:    updatedContent,
		}

		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactionByID(gomock.Any(), reactionID).
			Return(&existingReaction, nil).
			Times(1)
		h.repo.MockRollCallReactionRepository.EXPECT().
			UpdateRollCallReaction(gomock.Any(), reactionID, gomock.Any()).
			Return(nil).
			Times(1)
		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactionByID(gomock.Any(), reactionID).
			Return(&updatedReaction, nil).
			Times(1)

		h.expect.PUT("/api/reactions/{reactionId}", reactionID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(updateReqBody).
			Expect().
			Status(http.StatusOK)

		// SSEイベントの受信と検証 (Updated)
		if assert.True(t, scanner.Scan()) {
			line := scanner.Text()

			if assert.True(
				t,
				strings.HasPrefix(line, eventStreamDataPrefix),
				"line not start with 'data: '",
				line,
			) {
				assert.JSONEq(
					t,
					fmt.Sprintf(
						`{"id":%d,"type":"updated","userId":"%s","content":"%s"}`,
						reactionID,
						userID,
						updatedContent,
					),
					strings.TrimPrefix(line, eventStreamDataPrefix),
				)
			}
		}

		if assert.True(t, scanner.Scan()) {
			assert.Empty(t, scanner.Text())
		}

		// 3. Delete Reaction
		h.repo.MockRollCallReactionRepository.EXPECT().
			GetRollCallReactionByID(gomock.Any(), reactionID).
			Return(&updatedReaction, nil).
			Times(1)
		h.repo.MockRollCallReactionRepository.EXPECT().
			DeleteRollCallReaction(gomock.Any(), reactionID).
			Return(nil).
			Times(1)

		h.expect.DELETE("/api/reactions/{reactionId}", reactionID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusNoContent)

			// SSEイベントの受信と検証 (Deleted)
		if assert.True(t, scanner.Scan(), "should receive a line") {
			line := scanner.Text()

			if assert.True(
				t,
				strings.HasPrefix(line, eventStreamDataPrefix),
				"line not start with 'data: '",
				line,
			) {
				assert.JSONEq(
					t,
					fmt.Sprintf(`{"id":%d,"type":"deleted","userId":"%s"}`, reactionID, userID),
					strings.TrimPrefix(line, eventStreamDataPrefix),
				)
			}
		}

		if assert.True(t, scanner.Scan()) {
			assert.Empty(t, scanner.Text())
		}
	})

	t.Run("Filtering", func(t *testing.T) {
		t.Parallel()
		h := setup(t)
		rollCallID1 := uint(random.PositiveInt(t))
		rollCallID2 := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		user := model.User{ID: userID}

		// rollCallID1のストリームに接続
		ctx, cancel := context.WithTimeout(t.Context(), 1*time.Second)

		defer cancel()

		req := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			fmt.Sprintf("/api/roll-calls/%d/reactions/stream", rollCallID1),
			nil,
		)
		rec := httptest.NewRecorder()

		go func() {
			h.e.ServeHTTP(rec, req)
		}()

		// レスポンスヘッダーの確認
		assert.Eventually(t, func() bool {
			return rec.Header().Get(echo.HeaderContentType) == "text/event-stream"
		}, 500*time.Millisecond, 10*time.Millisecond, "content-type not text/event-stream", rec.Header().Get(echo.HeaderContentType))

		// rollCallID2のイベントを発生させる
		content := random.AlphaNumericString(t, 20)
		createReqBody := api.PostRollCallReactionJSONRequestBody{Content: content}
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&user, nil).
			Times(1)
		h.repo.MockRollCallReactionRepository.EXPECT().
			CreateRollCallReaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, reaction *model.RollCallReaction) error {
				reaction.ID = uint(random.PositiveInt(t))
				reaction.RollCallID = rollCallID2
				reaction.Content = content
				return nil
			}).
			Times(1)

		h.expect.POST("/api/roll-calls/{rollCallId}/reactions", rollCallID2).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(createReqBody).
			Expect().
			Status(http.StatusCreated)

		// ストリームにデータが流れてこないことを確認
		// コンテキストがタイムアウトするまで待機
		<-ctx.Done()

		// レスポンスボディに何も書き込まれていないことを確認
		body := rec.Body.String()
		// SSEヘッダーのみでイベントデータは含まれていないことを確認
		assert.Empty(t, body)
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		t.Parallel()
		h := setup(t)
		rollCallID := uint(random.PositiveInt(t))

		ctx, cancel := context.WithCancel(t.Context())
		req := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			fmt.Sprintf("/api/roll-calls/%d/reactions/stream", rollCallID),
			nil,
		)
		rec := httptest.NewRecorder()

		done := make(chan struct{})
		go func() {
			h.e.ServeHTTP(rec, req)
			close(done)
		}()

		// レスポンスヘッダーの確認
		assert.Eventually(t, func() bool {
			return rec.Header().Get(echo.HeaderContentType) == "text/event-stream"
		}, time.Second, 10*time.Millisecond, "content-type not text/event-stream", rec.Header().Get(echo.HeaderContentType))

		// コンテキストをキャンセル（クライアント切断をシミュレート）
		cancel()

		// ハンドラが正常に終了することを確認
		select {
		case <-done:
			// 正常に終了
		case <-time.After(2 * time.Second):
			t.Error("handler should finish when context is cancelled")
		}
	})
}

package router

import (
	"net/http"
	"testing"

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
}

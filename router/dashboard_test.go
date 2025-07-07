package router

import (
	"errors"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/traPtitech/rucQ/testutil/random"
)

func TestGetDashboard(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		userID := random.AlphaNumericString(t, 32)

		// モックの設定: ユーザーが参加者である場合
		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), uint(campID), userID).
			Return(true, nil)

		res := h.expect.GET("/api/camps/{campId}/me", campID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		res.Keys().ContainsOnly("id")
		res.Value("id").String().IsEqual(userID)
	})

	t.Run("Not Found - User is not a participant", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		userID := random.AlphaNumericString(t, 32)

		// モックの設定: ユーザーが参加者でない場合
		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), uint(campID), userID).
			Return(false, nil)

		h.expect.GET("/api/camps/{campId}/me", campID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			HasValue("message", "User is not a participant of this camp")
	})

	t.Run("Internal Server Error - Repository error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		userID := random.AlphaNumericString(t, 32)

		// モックの設定: リポジトリでエラーが発生する場合
		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), uint(campID), userID).
			Return(false, errors.New("database error"))

		h.expect.GET("/api/camps/{campId}/me", campID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusInternalServerError).
			JSON().
			Object().
			HasValue("message", "Internal server error")
	})
}

package router

import (
	"net/http"
	"testing"

	"github.com/traPtitech/rucQ/testutil/random"
)

func TestGetDashboard(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		userID := random.AlphaNumericString(t, 32)
		res := h.expect.GET("/api/camps/{campId}/me", campID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		res.Keys().ContainsOnly("id")
		res.Value("id").String().IsEqual(userID)
	})
}

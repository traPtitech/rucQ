package router

import (
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/traPtitech/rucQ/model"
	activityservice "github.com/traPtitech/rucQ/service/activity"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestGetActivities(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		user := &model.User{ID: userID}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(user, nil).
			Times(1)

		activities := []activityservice.ActivityResponse{
			{
				ID:   1,
				Type: model.ActivityTypeRoomCreated,
				Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				ID:   2,
				Type: model.ActivityTypePaymentAmountChanged,
				Time: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
				PaymentAmountChanged: &activityservice.PaymentChangedDetail{
					Amount: 1000,
				},
			},
		}

		h.activityService.EXPECT().
			GetActivities(gomock.Any(), campID, userID).
			Return(activities, nil).
			Times(1)

		res := h.expect.GET("/api/camps/{campId}/activities", campID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		res.Length().IsEqual(2)

		// Check first activity (RoomCreated)
		act1 := res.Value(0).Object()
		act1.Value("id").Number().IsEqual(1)
		act1.Value("type").String().IsEqual("room_created")
		act1.Value("time").String().IsEqual("2023-01-01T00:00:00Z")

		// Check second activity (PaymentAmountChanged)
		act2 := res.Value(1).Object()
		act2.Value("id").Number().IsEqual(2)
		act2.Value("type").String().IsEqual("payment_amount_changed")
		act2.Value("time").String().IsEqual("2023-01-02T00:00:00Z")
		act2.Value("amount").Number().IsEqual(1000)
	})

	t.Run("Unauthorized (missing header)", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))

		h.expect.GET("/api/camps/{campId}/activities", campID).
			Expect().
			Status(http.StatusBadRequest)
	})
}

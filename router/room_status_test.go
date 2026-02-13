package router

import (
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestServer_PutRoomStatus(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		roomID := api.RoomId(random.PositiveInt(t))
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		req := api.PutRoomStatusJSONRequestBody{
			Type:  api.RoomStatusTypeActive,
			Topic: "In meeting",
		}

		h.repo.MockRoomRepository.EXPECT().
			GetRoomCampID(gomock.Any(), uint(roomID)).
			Return(campID, nil).
			Times(1)
		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), campID, userID).
			Return(true, nil).
			Times(1)
		h.repo.MockRoomStatusRepository.EXPECT().
			SetRoomStatus(gomock.Any(), uint(roomID), gomock.Any(), userID).
			DoAndReturn(func(_ any, _ uint, status *model.RoomStatus, _ string) error {
				if status.Type != string(req.Type) {
					t.Fatalf("unexpected status type: %s", status.Type)
				}
				if status.Topic != req.Topic {
					t.Fatalf("unexpected status topic: %s", status.Topic)
				}
				return nil
			}).
			Times(1)

		h.expect.PUT("/api/rooms/{roomId}/status", roomID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(req).
			Expect().
			Status(http.StatusNoContent)
	})

	t.Run("Not Found - Room does not exist", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		roomID := api.RoomId(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		h.repo.MockRoomRepository.EXPECT().
			GetRoomCampID(gomock.Any(), uint(roomID)).
			Return(uint(0), repository.ErrRoomNotFound).
			Times(1)

		h.expect.PUT("/api/rooms/{roomId}/status", roomID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(api.PutRoomStatusJSONRequestBody{
				Type:  api.RoomStatusTypeActive,
				Topic: "Unavailable",
			}).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			HasValue("message", "Not Found")
	})

	t.Run("Not Found - User is not a participant", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		roomID := api.RoomId(random.PositiveInt(t))
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		h.repo.MockRoomRepository.EXPECT().
			GetRoomCampID(gomock.Any(), uint(roomID)).
			Return(campID, nil).
			Times(1)
		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), campID, userID).
			Return(false, nil).
			Times(1)

		h.expect.PUT("/api/rooms/{roomId}/status", roomID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(api.PutRoomStatusJSONRequestBody{
				Type:  api.RoomStatusTypeInactive,
				Topic: "Out",
			}).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			HasValue("message", "Not Found")
	})

	t.Run("Not Found - Camp does not exist", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		roomID := api.RoomId(random.PositiveInt(t))
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		h.repo.MockRoomRepository.EXPECT().
			GetRoomCampID(gomock.Any(), uint(roomID)).
			Return(campID, nil).
			Times(1)
		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), campID, userID).
			Return(false, model.ErrNotFound).
			Times(1)

		h.expect.PUT("/api/rooms/{roomId}/status", roomID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(api.PutRoomStatusJSONRequestBody{
				Type:  api.RoomStatusTypeInactive,
				Topic: "Out",
			}).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			HasValue("message", "Not Found")
	})
}

func TestServer_GetRoomStatusLogs(t *testing.T) {
	t.Parallel()

	t.Run("Success - Empty", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		roomID := api.RoomId(random.PositiveInt(t))
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		h.repo.MockRoomRepository.EXPECT().
			GetRoomCampID(gomock.Any(), uint(roomID)).
			Return(campID, nil).
			Times(1)
		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), campID, userID).
			Return(true, nil).
			Times(1)
		h.repo.MockRoomStatusRepository.EXPECT().
			GetRoomStatusLogs(gomock.Any(), uint(roomID)).
			Return([]model.RoomStatusLog{}, nil).
			Times(1)

		h.expect.GET("/api/rooms/{roomId}/status-logs", roomID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array().
			IsEmpty()
	})

	t.Run("Success - With Logs", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		roomID := api.RoomId(random.PositiveInt(t))
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		operatorID := random.AlphaNumericString(t, 32)
		now := time.Now().UTC().Truncate(time.Second)

		logs := []model.RoomStatusLog{
			{
				Type:       "active",
				Topic:      "Session",
				OperatorID: operatorID,
				Model:      gorm.Model{UpdatedAt: now},
			},
		}

		h.repo.MockRoomRepository.EXPECT().
			GetRoomCampID(gomock.Any(), uint(roomID)).
			Return(campID, nil).
			Times(1)
		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), campID, userID).
			Return(true, nil).
			Times(1)
		h.repo.MockRoomStatusRepository.EXPECT().
			GetRoomStatusLogs(gomock.Any(), uint(roomID)).
			Return(logs, nil).
			Times(1)

		res := h.expect.GET("/api/rooms/{roomId}/status-logs", roomID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		res.Length().IsEqual(1)
		res.Value(0).Object().
			HasValue("type", "active").
			HasValue("topic", "Session").
			HasValue("operatorId", operatorID).
			HasValue("updatedAt", now.Format(time.RFC3339))
	})

	t.Run("Not Found - Room does not exist", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		roomID := api.RoomId(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		h.repo.MockRoomRepository.EXPECT().
			GetRoomCampID(gomock.Any(), uint(roomID)).
			Return(uint(0), repository.ErrRoomNotFound).
			Times(1)

		h.expect.GET("/api/rooms/{roomId}/status-logs", roomID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			HasValue("message", "Not Found")
	})

	t.Run("Not Found - User is not a participant", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		roomID := api.RoomId(random.PositiveInt(t))
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		h.repo.MockRoomRepository.EXPECT().
			GetRoomCampID(gomock.Any(), uint(roomID)).
			Return(campID, nil).
			Times(1)
		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), campID, userID).
			Return(false, nil).
			Times(1)

		h.expect.GET("/api/rooms/{roomId}/status-logs", roomID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			HasValue("message", "Not Found")
	})

	t.Run("Not Found - Camp does not exist", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		roomID := api.RoomId(random.PositiveInt(t))
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		h.repo.MockRoomRepository.EXPECT().
			GetRoomCampID(gomock.Any(), uint(roomID)).
			Return(campID, nil).
			Times(1)
		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), campID, userID).
			Return(false, model.ErrNotFound).
			Times(1)

		h.expect.GET("/api/rooms/{roomId}/status-logs", roomID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			HasValue("message", "Not Found")
	})
}

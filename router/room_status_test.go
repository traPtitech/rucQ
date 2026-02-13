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

	t.Run("成功", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		roomID := api.RoomId(random.PositiveInt(t))
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		statusType := random.SelectFrom(t, api.RoomStatusTypeActive, api.RoomStatusTypeInactive)

		req := api.PutRoomStatusJSONRequestBody{
			Type:  statusType,
			Topic: random.AlphaNumericString(t, 64),
		}

		h.repo.MockRoomRepository.EXPECT().
			GetRoomCampID(gomock.Any(), uint(roomID)).
			Return(campID, nil).
			Times(1)
		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), campID, userID).
			Return(true, nil).
			Times(1)
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&model.User{ID: userID}, nil).
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

	t.Run("部屋が存在しない", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		roomID := api.RoomId(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		h.repo.MockRoomRepository.EXPECT().
			GetRoomCampID(gomock.Any(), uint(roomID)).
			Return(uint(0), repository.ErrRoomNotFound).
			Times(1)

		statusType := random.SelectFrom(t, api.RoomStatusTypeActive, api.RoomStatusTypeInactive)

		h.expect.PUT("/api/rooms/{roomId}/status", roomID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(api.PutRoomStatusJSONRequestBody{
				Type:  statusType,
				Topic: random.AlphaNumericString(t, 64),
			}).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			HasValue("message", "Not Found")
	})

	t.Run("ユーザーが合宿の参加者ではない", func(t *testing.T) {
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
				Topic: random.AlphaNumericString(t, 64),
			}).
			Expect().
			Status(http.StatusForbidden).
			JSON().
			Object().
			HasValue("message", "Forbidden")
	})

	t.Run("合宿が存在しない", func(t *testing.T) {
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
			Return(false, repository.ErrCampNotFound).
			Times(1)

		h.expect.PUT("/api/rooms/{roomId}/status", roomID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(api.PutRoomStatusJSONRequestBody{
				Type:  api.RoomStatusTypeInactive,
				Topic: random.AlphaNumericString(t, 64),
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

	t.Run("空のログ", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		roomID := api.RoomId(random.PositiveInt(t))

		h.repo.MockRoomStatusRepository.EXPECT().
			GetRoomStatusLogs(gomock.Any(), uint(roomID)).
			Return([]model.RoomStatusLog{}, nil).
			Times(1)

		h.expect.GET("/api/rooms/{roomId}/status-logs", roomID).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array().
			IsEmpty()
	})

	t.Run("複数の要素を含むログ", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		roomID := api.RoomId(random.PositiveInt(t))
		operatorID := random.AlphaNumericString(t, 32)
		updatedAt := random.Time(t)
		topic := random.AlphaNumericString(t, 64)
		statusType := random.SelectFrom(t, "active", "inactive")

		logs := []model.RoomStatusLog{
			{
				Type:       statusType,
				Topic:      topic,
				OperatorID: operatorID,
				Model:      gorm.Model{UpdatedAt: updatedAt},
			},
		}

		h.repo.MockRoomStatusRepository.EXPECT().
			GetRoomStatusLogs(gomock.Any(), uint(roomID)).
			Return(logs, nil).
			Times(1)

		res := h.expect.GET("/api/rooms/{roomId}/status-logs", roomID).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		res.Length().IsEqual(1)
		res.Value(0).Object().
			HasValue("type", statusType).
			HasValue("topic", topic).
			HasValue("operatorId", operatorID).
			HasValue("updatedAt", updatedAt.Format(time.RFC3339))
	})
}

package gormrepository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestRepository_SetRoomStatus(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})

		operatorID := random.AlphaNumericString(t, 32)

		status := &model.RoomStatus{
			Type:  "active",
			Topic: "Session",
		}

		if !assert.NoError(t, r.SetRoomStatus(t.Context(), room.ID, status, operatorID)) {
			return
		}

		retrievedRoom, err := r.GetRoomByID(t.Context(), room.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrievedRoom.Status)
		assert.Equal(t, status.Type, retrievedRoom.Status.Type)
		assert.Equal(t, status.Topic, retrievedRoom.Status.Topic)

		updatedStatus := &model.RoomStatus{
			Type:  "inactive",
			Topic: "Break",
		}

		time.Sleep(10 * time.Millisecond)
		if !assert.NoError(t, r.SetRoomStatus(t.Context(), room.ID, updatedStatus, operatorID)) {
			return
		}

		retrievedRoom, err = r.GetRoomByID(t.Context(), room.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrievedRoom.Status)
		assert.Equal(t, updatedStatus.Type, retrievedRoom.Status.Type)
		assert.Equal(t, updatedStatus.Topic, retrievedRoom.Status.Topic)

		logs, err := r.GetRoomStatusLogs(t.Context(), room.ID)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, logs, 2)
		assert.Equal(t, "active", logs[0].Type)
		assert.Equal(t, "inactive", logs[1].Type)
	})

	t.Run("RoomNotFound", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		err := r.SetRoomStatus(t.Context(), uint(random.PositiveInt(t)), &model.RoomStatus{
			Type:  "active",
			Topic: "Session",
		}, random.AlphaNumericString(t, 32))

		assert.ErrorIs(t, err, repository.ErrRoomNotFound)
	})
}

func TestRepository_GetRoomStatusLogs(t *testing.T) {
	t.Parallel()

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})

		logs, err := r.GetRoomStatusLogs(t.Context(), room.ID)
		if !assert.NoError(t, err) {
			return
		}
		assert.Empty(t, logs)
	})
}

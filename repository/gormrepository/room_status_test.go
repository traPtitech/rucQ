package gormrepository

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestRepository_SetRoomStatus(t *testing.T) {
	t.Parallel()

	t.Run("Create", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})

		operator := mustCreateUser(t, r)
		operatorID := operator.ID
		topic := random.AlphaNumericString(t, 64)

		status := &model.RoomStatus{
			Type:  "active",
			Topic: topic,
		}

		err := r.SetRoomStatus(t.Context(), room.ID, status, operatorID)
		assert.NoError(t, err)

		retrievedRoom, err := r.GetRoomByID(t.Context(), room.ID)
		assert.NoError(t, err)
		if assert.NotNil(t, retrievedRoom) {
			assert.NotNil(t, retrievedRoom.Status)
			if retrievedRoom.Status != nil {
				assert.Equal(t, status.Type, retrievedRoom.Status.Type)
				assert.Equal(t, status.Topic, retrievedRoom.Status.Topic)
			}
		}
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})

		operator := mustCreateUser(t, r)
		operatorID := operator.ID
		initialTopic := random.AlphaNumericString(t, 64)
		mustSetRoomStatus(t, r, room.ID, &model.RoomStatus{
			Type:  "active",
			Topic: initialTopic,
		}, operatorID)

		updatedTopic := random.AlphaNumericString(t, 64)
		updatedStatus := &model.RoomStatus{
			Type:  "inactive",
			Topic: updatedTopic,
		}

		err := r.SetRoomStatus(t.Context(), room.ID, updatedStatus, operatorID)
		assert.NoError(t, err)

		retrievedRoom, err := r.GetRoomByID(t.Context(), room.ID)
		assert.NoError(t, err)
		if assert.NotNil(t, retrievedRoom) {
			assert.NotNil(t, retrievedRoom.Status)
			if retrievedRoom.Status != nil {
				assert.Equal(t, updatedStatus.Type, retrievedRoom.Status.Type)
				assert.Equal(t, updatedStatus.Topic, retrievedRoom.Status.Topic)
			}
		}

		logs, err := r.GetRoomStatusLogs(t.Context(), room.ID)
		assert.NoError(t, err)
		assert.Len(t, logs, 2)
		if len(logs) >= 2 {
			assert.Equal(t, "inactive", logs[0].Type)
			assert.Equal(t, "active", logs[1].Type)
		}
	})

	t.Run("JapaneseTopic64", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})

		operator := mustCreateUser(t, r)
		operatorID := operator.ID
		japaneseTopic := strings.Repeat("あ", 64)

		status := &model.RoomStatus{
			Type:  "active",
			Topic: japaneseTopic,
		}

		err := r.SetRoomStatus(t.Context(), room.ID, status, operatorID)
		assert.NoError(t, err)

		retrievedRoom, err := r.GetRoomByID(t.Context(), room.ID)
		assert.NoError(t, err)
		if assert.NotNil(t, retrievedRoom) {
			assert.NotNil(t, retrievedRoom.Status)
			if retrievedRoom.Status != nil {
				assert.Equal(t, status.Topic, retrievedRoom.Status.Topic)
			}
		}
	})

	t.Run("RoomNotFound", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		err := r.SetRoomStatus(t.Context(), uint(random.PositiveInt(t)), &model.RoomStatus{
			Type:  "active",
			Topic: random.AlphaNumericString(t, 64),
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
		assert.NoError(t, err)
		assert.Empty(t, logs)
	})

	t.Run("Multiple", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})
		operator := mustCreateUser(t, r)
		operatorID := operator.ID

		mustSetRoomStatus(t, r, room.ID, &model.RoomStatus{
			Type:  "active",
			Topic: random.AlphaNumericString(t, 64),
		}, operatorID)

		mustSetRoomStatus(t, r, room.ID, &model.RoomStatus{
			Type:  "inactive",
			Topic: random.AlphaNumericString(t, 64),
		}, operatorID)

		logs, err := r.GetRoomStatusLogs(t.Context(), room.ID)
		assert.NoError(t, err)
		assert.Len(t, logs, 2)
		if len(logs) >= 2 {
			assert.Equal(t, "inactive", logs[0].Type)
			assert.Equal(t, "active", logs[1].Type)
		}
	})
}

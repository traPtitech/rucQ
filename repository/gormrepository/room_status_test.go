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

	const roomStatusTopicMaxLength = 64

	t.Run("作成", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})

		operator := mustCreateUser(t, r)
		operatorID := operator.ID
		topic := random.AlphaNumericString(t, roomStatusTopicMaxLength)
		statusType := random.SelectFrom(t, "active", "inactive")

		status := model.RoomStatus{
			Type:  &statusType,
			Topic: topic,
		}

		err := r.SetRoomStatus(t.Context(), room.ID, status, operatorID)
		assert.NoError(t, err)

		retrievedRoom, err := r.GetRoomByID(t.Context(), room.ID)
		assert.NoError(t, err)

		if assert.NotNil(t, retrievedRoom) {
			assert.Equal(t, status.Type, retrievedRoom.Status.Type)
			assert.Equal(t, status.Topic, retrievedRoom.Status.Topic)
		}
	})

	t.Run("更新", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})

		operator := mustCreateUser(t, r)
		operatorID := operator.ID
		initialTopic := random.AlphaNumericString(t, roomStatusTopicMaxLength)
		statusTypeOld := random.SelectFrom(t, "active", "inactive")
		statusTypeNew := random.SelectFrom(t, "active", "inactive")

		mustSetRoomStatus(t, r, room.ID, model.RoomStatus{
			Type:  &statusTypeOld,
			Topic: initialTopic,
		}, operatorID)

		updatedTopic := random.AlphaNumericString(t, roomStatusTopicMaxLength)
		updatedStatus := model.RoomStatus{
			Type:  &statusTypeNew,
			Topic: updatedTopic,
		}

		err := r.SetRoomStatus(t.Context(), room.ID, updatedStatus, operatorID)
		assert.NoError(t, err)

		retrievedRoom, err := r.GetRoomByID(t.Context(), room.ID)
		assert.NoError(t, err)

		if assert.NotNil(t, retrievedRoom) {
			assert.Equal(t, updatedStatus.Type, retrievedRoom.Status.Type)
			assert.Equal(t, updatedStatus.Topic, retrievedRoom.Status.Topic)
		}

		logs, err := r.GetRoomStatusLogs(t.Context(), room.ID)
		assert.NoError(t, err)

		if assert.Len(t, logs, 2) &&
			assert.NotNil(t, logs[0].Type) &&
			assert.NotNil(t, logs[1].Type) {
			assert.Equal(t, statusTypeOld, *logs[0].Type)
			assert.Equal(t, statusTypeNew, *logs[1].Type)
		}
	})

	t.Run("nullで上書きできる", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})

		operator := mustCreateUser(t, r)
		operatorID := operator.ID
		statusType := random.SelectFrom(t, "active", "inactive")

		mustSetRoomStatus(t, r, room.ID, model.RoomStatus{
			Type:  &statusType,
			Topic: random.AlphaNumericString(t, roomStatusTopicMaxLength),
		}, operatorID)

		updatedTopic := random.AlphaNumericString(t, roomStatusTopicMaxLength)
		err := r.SetRoomStatus(t.Context(), room.ID, model.RoomStatus{
			Type:  nil,
			Topic: updatedTopic,
		}, operatorID)
		assert.NoError(t, err)

		retrievedRoom, err := r.GetRoomByID(t.Context(), room.ID)
		assert.NoError(t, err)

		if assert.NotNil(t, retrievedRoom) {
			assert.Nil(t, retrievedRoom.Status.Type)
			assert.Equal(t, updatedTopic, retrievedRoom.Status.Topic)
		}
	})

	t.Run("マルチバイト文字でも64文字までは保存できる", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})

		operator := mustCreateUser(t, r)
		operatorID := operator.ID
		japaneseTopic := strings.Repeat("あ", roomStatusTopicMaxLength)
		statusType := random.SelectFrom(t, "active", "inactive")
		status := model.RoomStatus{
			Type:  &statusType,
			Topic: japaneseTopic,
		}

		err := r.SetRoomStatus(t.Context(), room.ID, status, operatorID)
		assert.NoError(t, err)

		retrievedRoom, err := r.GetRoomByID(t.Context(), room.ID)
		assert.NoError(t, err)

		if assert.NotNil(t, retrievedRoom) {
			assert.Equal(t, status.Topic, retrievedRoom.Status.Topic)
		}
	})

	t.Run("部屋が存在しない", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		statusType := random.SelectFrom(t, "active", "inactive")
		err := r.SetRoomStatus(t.Context(), uint(random.PositiveInt(t)), model.RoomStatus{
			Type:  &statusType,
			Topic: random.AlphaNumericString(t, roomStatusTopicMaxLength),
		}, random.AlphaNumericString(t, 32))

		assert.ErrorIs(t, err, repository.ErrRoomNotFound)
	})

	t.Run("64文字を超えるとエラー", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})

		operator := mustCreateUser(t, r)
		operatorID := operator.ID
		statusType := random.SelectFrom(t, "active", "inactive")
		err := r.SetRoomStatus(t.Context(), room.ID, model.RoomStatus{
			Type:  &statusType,
			Topic: strings.Repeat("a", roomStatusTopicMaxLength+1),
		}, operatorID)

		assert.Error(t, err)
	})
}

func TestRepository_GetRoomStatusLogs(t *testing.T) {
	t.Parallel()

	t.Run("空のログ", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})

		logs, err := r.GetRoomStatusLogs(t.Context(), room.ID)
		assert.NoError(t, err)
		assert.Empty(t, logs)
	})

	t.Run("複数の要素を含むログ", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})
		operator := mustCreateUser(t, r)
		operatorID := operator.ID
		statusTypeOld := random.SelectFrom(t, "active", "inactive")
		statusTypeNew := random.SelectFrom(t, "active", "inactive")

		mustSetRoomStatus(t, r, room.ID, model.RoomStatus{
			Type:  &statusTypeOld,
			Topic: random.AlphaNumericString(t, 64),
		}, operatorID)

		mustSetRoomStatus(t, r, room.ID, model.RoomStatus{
			Type:  &statusTypeNew,
			Topic: random.AlphaNumericString(t, 64),
		}, operatorID)

		logs, err := r.GetRoomStatusLogs(t.Context(), room.ID)
		assert.NoError(t, err)

		if assert.Len(t, logs, 2) &&
			assert.NotNil(t, logs[0].Type) &&
			assert.NotNil(t, logs[1].Type) {
			// 新しい順で返ってくることを確認
			assert.Equal(t, statusTypeNew, *logs[0].Type)
			assert.Equal(t, statusTypeOld, *logs[1].Type)
		}
	})

	t.Run("部屋が存在しない", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		logs, err := r.GetRoomStatusLogs(t.Context(), uint(random.PositiveInt(t)))
		assert.ErrorIs(t, err, repository.ErrRoomNotFound)
		assert.Nil(t, logs)
	})
}

package gormrepository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestRepository_GetRoomCampID(t *testing.T) {
	t.Parallel()

	t.Run("成功", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})

		campID, err := r.GetRoomCampID(t.Context(), room.ID)
		assert.NoError(t, err)
		assert.Equal(t, camp.ID, campID)
	})

	t.Run("部屋が存在しない", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		_, err := r.GetRoomCampID(t.Context(), uint(random.PositiveInt(t)))
		assert.ErrorIs(t, err, repository.ErrRoomNotFound)
	})
}

func TestRepository_GetRoomByUserID(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		user1 := mustCreateUser(t, r)
		user2 := mustCreateUser(t, r)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{user1, user2})

		retrievedRoom, err := r.GetRoomByUserID(t.Context(), camp.ID, user1.ID)

		assert.NoError(t, err)
		assert.Equal(t, room.ID, retrievedRoom.ID)
		assert.Equal(t, room.Name, retrievedRoom.Name)
		assert.Equal(t, room.RoomGroupID, retrievedRoom.RoomGroupID)

		if assert.Len(t, retrievedRoom.Members, 2) {
			memberIDs := make([]string, len(retrievedRoom.Members))

			for i, member := range retrievedRoom.Members {
				memberIDs[i] = member.ID
			}

			assert.Contains(t, memberIDs, user1.ID)
			assert.Contains(t, memberIDs, user2.ID)
		}
	})

	t.Run("最新のステータスが含まれる", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		user := mustCreateUser(t, r)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{user})
		operator := mustCreateUser(t, r)
		statusTypeOld := random.SelectFrom(t, "active", "inactive")
		statusTypeNew := random.SelectFrom(t, "active", "inactive")

		mustSetRoomStatus(t, r, room.ID, model.RoomStatus{
			Type:  &statusTypeOld,
			Topic: random.AlphaNumericString(t, 64),
		}, operator.ID)

		latestTopic := random.AlphaNumericString(t, 64)
		mustSetRoomStatus(t, r, room.ID, model.RoomStatus{
			Type:  &statusTypeNew,
			Topic: latestTopic,
		}, operator.ID)

		retrievedRoom, err := r.GetRoomByUserID(t.Context(), camp.ID, user.ID)
		assert.NoError(t, err)

		if assert.NotNil(t, retrievedRoom) && assert.NotNil(t, retrievedRoom.Status.Type) {
			assert.Equal(t, statusTypeNew, *retrievedRoom.Status.Type)
			assert.Equal(t, latestTopic, retrievedRoom.Status.Topic)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		// ユーザーは作成するが部屋には所属させない
		user := mustCreateUser(t, r)
		_, err := r.GetRoomByUserID(t.Context(), camp.ID, user.ID)

		assert.ErrorIs(t, err, repository.ErrRoomNotFound)
	})
}

func TestRepository_GetRoomByID(t *testing.T) {
	t.Parallel()

	t.Run("最新のステータスが含まれる", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})
		operator := mustCreateUser(t, r)
		statusTypeOld := random.SelectFrom(t, "active", "inactive")
		statusTypeNew := random.SelectFrom(t, "active", "inactive")

		mustSetRoomStatus(t, r, room.ID, model.RoomStatus{
			Type:  &statusTypeOld,
			Topic: random.AlphaNumericString(t, 64),
		}, operator.ID)

		latestTopic := random.AlphaNumericString(t, 64)
		mustSetRoomStatus(t, r, room.ID, model.RoomStatus{
			Type:  &statusTypeNew,
			Topic: latestTopic,
		}, operator.ID)

		retrievedRoom, err := r.GetRoomByID(t.Context(), room.ID)
		assert.NoError(t, err)

		if assert.NotNil(t, retrievedRoom) && assert.NotNil(t, retrievedRoom.Status.Type) {
			assert.Equal(t, statusTypeNew, *retrievedRoom.Status.Type)
			assert.Equal(t, latestTopic, retrievedRoom.Status.Topic)
		}
	})
}

func TestRepository_GetRooms(t *testing.T) {
	t.Parallel()

	t.Run("最新のステータスが含まれる", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})
		operator := mustCreateUser(t, r)
		statusTypeOld := random.SelectFrom(t, "active", "inactive")
		statusTypeNew := random.SelectFrom(t, "active", "inactive")

		mustSetRoomStatus(t, r, room.ID, model.RoomStatus{
			Type:  &statusTypeOld,
			Topic: random.AlphaNumericString(t, 64),
		}, operator.ID)

		latestTopic := random.AlphaNumericString(t, 64)
		mustSetRoomStatus(t, r, room.ID, model.RoomStatus{
			Type:  &statusTypeNew,
			Topic: latestTopic,
		}, operator.ID)

		rooms, err := r.GetRooms()
		assert.NoError(t, err)

		if assert.Len(t, rooms, 1) && assert.NotNil(t, rooms[0].Status.Type) {
			assert.Equal(t, "inactive", *rooms[0].Status.Type)
			assert.Equal(t, latestTopic, rooms[0].Status.Topic)
		}
	})
}

func TestRepository_CreateRoom(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		user1 := mustCreateUser(t, r)
		user2 := mustCreateUser(t, r)
		members := []model.User{user1, user2}

		room := &model.Room{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupID: roomGroup.ID,
			Members:     members,
		}

		err := r.CreateRoom(t.Context(), room)

		assert.NoError(t, err)
		assert.NotZero(t, room.ID)
		assert.Equal(t, roomGroup.ID, room.RoomGroupID)

		// 作成された部屋を取得して確認
		retrievedRoom, err := r.GetRoomByID(t.Context(), room.ID)

		require.NoError(t, err)
		assert.Equal(t, room.Name, retrievedRoom.Name)
		assert.Equal(t, roomGroup.ID, retrievedRoom.RoomGroupID)

		if assert.Len(t, retrievedRoom.Members, 2) {
			// メンバーのIDが正しく設定されているか確認
			memberIDs := make([]string, len(retrievedRoom.Members))

			for i, member := range retrievedRoom.Members {
				memberIDs[i] = member.ID
			}

			assert.Contains(t, memberIDs, user1.ID)
			assert.Contains(t, memberIDs, user2.ID)
		}

		assert.Equal(t, "", retrievedRoom.Status.Topic)
		assert.Nil(t, retrievedRoom.Status.Type)
	})

	t.Run("Success without members", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)

		room := &model.Room{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupID: roomGroup.ID,
			Members:     []model.User{},
		}

		err := r.CreateRoom(t.Context(), room)

		assert.NoError(t, err)
		assert.NotZero(t, room.ID)

		// 作成された部屋を取得して確認
		retrievedRoom, err := r.GetRoomByID(t.Context(), room.ID)

		require.NoError(t, err)
		assert.Equal(t, room.Name, retrievedRoom.Name)
		assert.Equal(t, roomGroup.ID, retrievedRoom.RoomGroupID)
		assert.Empty(t, retrievedRoom.Members)
		assert.Equal(t, "", retrievedRoom.Status.Topic)
		assert.Nil(t, retrievedRoom.Status.Type)
	})

	t.Run("Non-existent RoomGroup", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		user := mustCreateUser(t, r)
		nonExistentRoomGroupID := uint(random.PositiveInt(t))

		room := &model.Room{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupID: nonExistentRoomGroupID, // 存在しないRoomGroup
			Members:     []model.User{user},
		}

		err := r.CreateRoom(t.Context(), room)

		assert.ErrorIs(t, err, repository.ErrUserOrRoomGroupNotFound)
	})

	t.Run("Non-existent User", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)

		// 存在しないユーザーをメンバーに設定
		nonExistentUser := model.User{ID: random.AlphaNumericString(t, 32)}
		room := &model.Room{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupID: roomGroup.ID,
			Members:     []model.User{nonExistentUser},
		}

		err := r.CreateRoom(t.Context(), room)

		assert.ErrorIs(t, err, repository.ErrUserOrRoomGroupNotFound)
	})

	t.Run("Mixed existing and non-existent users", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		existingUser := mustCreateUser(t, r)
		nonExistentUser := model.User{ID: random.AlphaNumericString(t, 32)}
		room := &model.Room{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupID: roomGroup.ID,
			Members:     []model.User{existingUser, nonExistentUser},
		}

		err := r.CreateRoom(t.Context(), room)

		assert.ErrorIs(t, err, repository.ErrUserOrRoomGroupNotFound)
	})
}

func TestRepository_UpdateRoom(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		user1 := mustCreateUser(t, r)
		user2 := mustCreateUser(t, r)
		originalMembers := []model.User{user1}
		// 元の部屋を作成
		room := mustCreateRoom(t, r, roomGroup.ID, originalMembers)
		// 新しい情報で更新
		newName := random.AlphaNumericString(t, 25)
		updatedMembers := []model.User{user1, user2}
		updatedRoom := &model.Room{
			Name:        newName,
			RoomGroupID: roomGroup.ID,
			Members:     updatedMembers,
		}

		err := r.UpdateRoom(t.Context(), room.ID, updatedRoom)

		assert.NoError(t, err)

		// 更新が正しく反映されているか確認
		retrievedRoom, err := r.GetRoomByID(t.Context(), room.ID)

		require.NoError(t, err)
		assert.Equal(t, newName, retrievedRoom.Name)
		assert.Equal(t, roomGroup.ID, retrievedRoom.RoomGroupID)

		if assert.Len(t, retrievedRoom.Members, 2) {
			// メンバーのIDが正しく設定されているか確認
			memberIDs := make([]string, len(retrievedRoom.Members))

			for i, member := range retrievedRoom.Members {
				memberIDs[i] = member.ID
			}

			assert.Contains(t, memberIDs, user1.ID)
			assert.Contains(t, memberIDs, user2.ID)
		}
	})

	t.Run("ステータスは上書きされない", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})
		operator := mustCreateUser(t, r)

		latestTopic := random.AlphaNumericString(t, 64)
		statusType := random.SelectFrom(t, "active", "inactive")

		mustSetRoomStatus(t, r, room.ID, model.RoomStatus{
			Type:  &statusType,
			Topic: latestTopic,
		}, operator.ID)

		updatedRoom := &model.Room{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupID: roomGroup.ID,
			Members:     []model.User{},
		}

		err := r.UpdateRoom(t.Context(), room.ID, updatedRoom)
		assert.NoError(t, err)

		retrievedRoom, err := r.GetRoomByID(t.Context(), room.ID)
		assert.NoError(t, err)

		if assert.NotNil(t, retrievedRoom) && assert.NotNil(t, retrievedRoom.Status.Type) {
			assert.Equal(t, statusType, *retrievedRoom.Status.Type)
			assert.Equal(t, latestTopic, retrievedRoom.Status.Topic)
		}
	})

	t.Run("Success - Remove all members", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		user1 := mustCreateUser(t, r)
		user2 := mustCreateUser(t, r)
		originalMembers := []model.User{user1, user2}
		// 元の部屋を作成
		room := mustCreateRoom(t, r, roomGroup.ID, originalMembers)
		// メンバーを空にして更新
		updatedRoom := &model.Room{
			Name:        room.Name,
			RoomGroupID: roomGroup.ID,
			Members:     []model.User{},
		}
		err := r.UpdateRoom(t.Context(), room.ID, updatedRoom)

		assert.NoError(t, err)

		// 更新が正しく反映されているか確認
		retrievedRoom, err := r.GetRoomByID(t.Context(), room.ID)

		require.NoError(t, err)
		assert.Equal(t, room.Name, retrievedRoom.Name)
		assert.Empty(t, retrievedRoom.Members)
	})

	t.Run("Success - Change RoomGroup", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		originalRoomGroup := mustCreateRoomGroup(t, r, camp.ID)
		newRoomGroup := mustCreateRoomGroup(t, r, camp.ID)
		user := mustCreateUser(t, r)
		// 元の部屋を作成
		room := mustCreateRoom(t, r, originalRoomGroup.ID, []model.User{user})
		// 異なるRoomGroupに変更
		updatedRoom := &model.Room{
			Name:        room.Name,
			RoomGroupID: newRoomGroup.ID,
			Members:     []model.User{user},
		}
		err := r.UpdateRoom(t.Context(), room.ID, updatedRoom)

		assert.NoError(t, err)

		// 更新が正しく反映されているか確認
		retrievedRoom, err := r.GetRoomByID(t.Context(), room.ID)

		require.NoError(t, err)
		assert.Equal(t, newRoomGroup.ID, retrievedRoom.RoomGroupID)
	})

	t.Run("Non-existent Room", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		user := mustCreateUser(t, r)
		nonExistentRoomID := uint(random.PositiveInt(t))
		updatedRoom := &model.Room{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupID: roomGroup.ID,
			Members:     []model.User{user},
		}
		err := r.UpdateRoom(t.Context(), nonExistentRoomID, updatedRoom)

		assert.ErrorIs(t, err, repository.ErrRoomNotFound)
	})

	t.Run("Non-existent RoomGroup", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		user := mustCreateUser(t, r)
		// 元の部屋を作成
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{user})
		// 存在しないRoomGroupで更新
		nonExistentRoomGroupID := uint(random.PositiveInt(t))
		updatedRoom := &model.Room{
			Name:        room.Name,
			RoomGroupID: nonExistentRoomGroupID,
			Members:     []model.User{user},
		}
		err := r.UpdateRoom(t.Context(), room.ID, updatedRoom)

		assert.ErrorIs(t, err, repository.ErrRoomGroupNotFound)
	})

	t.Run("Non-existent User in Members", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		user := mustCreateUser(t, r)
		// 元の部屋を作成
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{user})
		// 存在しないユーザーをメンバーに追加して更新
		nonExistentUser := model.User{ID: random.AlphaNumericString(t, 32)}
		updatedRoom := &model.Room{
			Name:        room.Name,
			RoomGroupID: roomGroup.ID,
			Members:     []model.User{user, nonExistentUser},
		}

		err := r.UpdateRoom(t.Context(), room.ID, updatedRoom)

		assert.ErrorIs(t, err, repository.ErrUserNotFound)
	})

	t.Run("Failure - User already assigned in same camp", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		user := mustCreateUser(t, r)

		// 部屋1を作成し、ユーザーを所属させる
		_ = mustCreateRoom(t, r, roomGroup.ID, []model.User{user})

		// 部屋2を作成
		room2 := mustCreateRoom(t, r, roomGroup.ID, []model.User{})

		// 同じユーザーを部屋2にも追加しようとする
		updatedRoom := &model.Room{
			Name:        room2.Name,
			RoomGroupID: roomGroup.ID,
			Members:     []model.User{user},
		}

		err := r.UpdateRoom(t.Context(), room2.ID, updatedRoom)

		assert.ErrorIs(t, err, repository.ErrUserAlreadyAssigned)
	})

	t.Run("Success - User assigned in different camp", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		// 合宿A と 部屋1
		campA := mustCreateCamp(t, r)
		groupA := mustCreateRoomGroup(t, r, campA.ID)
		user := mustCreateUser(t, r)
		_ = mustCreateRoom(t, r, groupA.ID, []model.User{user})

		// 合宿B と 部屋2 (空)
		campB := mustCreateCamp(t, r)
		groupB := mustCreateRoomGroup(t, r, campB.ID)
		room2 := mustCreateRoom(t, r, groupB.ID, []model.User{})

		// 合宿Aにいるユーザーを、合宿Bの部屋にも追加する
		updatedRoom := &model.Room{
			Name:        room2.Name,
			RoomGroupID: groupB.ID,
			Members:     []model.User{user},
		}

		err := r.UpdateRoom(t.Context(), room2.ID, updatedRoom)

		// 合宿が異なるので、成功（NoError）を期待
		assert.NoError(t, err)

		// 実際に所属できているか確認
		retrievedRoom, err := r.GetRoomByID(t.Context(), room2.ID)
		require.NoError(t, err)
		assert.Len(t, retrievedRoom.Members, 1)
		assert.Equal(t, user.ID, retrievedRoom.Members[0].ID)
	})
}

func TestRepository_DeleteRoom(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		user := mustCreateUser(t, r)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{user})

		err := r.DeleteRoom(t.Context(), room.ID)

		assert.NoError(t, err)

		// 削除されているかを確認
		_, err = r.GetRoomByID(t.Context(), room.ID)
		assert.ErrorIs(t, err, repository.ErrRoomNotFound)
	})

	t.Run("ステータスも削除される", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})
		operator := mustCreateUser(t, r)
		statusType1 := random.SelectFrom(t, "active", "inactive")
		statusType2 := random.SelectFrom(t, "active", "inactive")

		mustSetRoomStatus(t, r, room.ID, model.RoomStatus{
			Type:  &statusType1,
			Topic: random.AlphaNumericString(t, 64),
		}, operator.ID)

		mustSetRoomStatus(t, r, room.ID, model.RoomStatus{
			Type:  &statusType2,
			Topic: random.AlphaNumericString(t, 64),
		}, operator.ID)

		err := r.DeleteRoom(t.Context(), room.ID)
		assert.NoError(t, err)

		var statusCount int64
		err = r.db.WithContext(t.Context()).
			Model(model.RoomStatus{}).
			Where("room_id = ?", room.ID).
			Count(&statusCount).Error
		assert.NoError(t, err)
		assert.Zero(t, statusCount)

		var logCount int64
		err = r.db.WithContext(t.Context()).
			Model(&model.RoomStatusLog{}).
			Where("room_id = ?", room.ID).
			Count(&logCount).Error
		assert.NoError(t, err)
		assert.Zero(t, logCount)
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		err := r.DeleteRoom(t.Context(), uint(random.PositiveInt(t)))

		assert.ErrorIs(t, err, repository.ErrRoomNotFound)
	})
}

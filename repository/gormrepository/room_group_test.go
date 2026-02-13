package gormrepository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestRepository_CreateRoomGroup(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := &model.RoomGroup{
			Name:   random.AlphaNumericString(t, 20),
			CampID: camp.ID,
		}

		err := r.CreateRoomGroup(t.Context(), roomGroup)

		assert.NoError(t, err)
		assert.NotZero(t, roomGroup.ID)
		assert.Equal(t, camp.ID, roomGroup.CampID)
	})

	t.Run("Empty Name", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := &model.RoomGroup{
			Name:   "",
			CampID: camp.ID,
		}

		err := r.CreateRoomGroup(t.Context(), roomGroup)

		assert.NoError(t, err) // 空の名前でもエラーにならないことを確認
		assert.NotZero(t, roomGroup.ID)
	})

	t.Run("Invalid campID", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		campID := uint(random.PositiveInt(t)) // 存在しないCampID
		roomGroup := &model.RoomGroup{
			Name:   random.AlphaNumericString(t, 20),
			CampID: campID,
		}

		err := r.CreateRoomGroup(t.Context(), roomGroup)

		assert.ErrorIs(t, err, repository.ErrCampNotFound)
	})

	t.Run("Failure - Duplicate users within the same request", func(t *testing.T) {
		t.Parallel()
		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)

		// 1つ目のリクエスト内に、同じユーザーを2つの部屋に入れてみる
		rg := &model.RoomGroup{
			CampID: camp.ID,
			Rooms: []model.Room{
				{Name: "Room A", Members: []model.User{user}},
				{Name: "Room B", Members: []model.User{user}}, // 重複！
			},
		}

		err := r.CreateRoomGroup(t.Context(), rg)
		assert.ErrorIs(t, err, repository.ErrUserAlreadyAssigned)
	})

	t.Run("Failure - User already exists in another RoomGroup of same camp", func(t *testing.T) {
		t.Parallel()
		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)

		rgA := &model.RoomGroup{
			Name:   "Existing Group",
			CampID: camp.ID,
			Rooms: []model.Room{
				{Name: "Occupied Room", Members: []model.User{user}},
			},
		}
		require.NoError(t, r.CreateRoomGroup(t.Context(), rgA))

		rgB := &model.RoomGroup{
			Name:   "New Group",
			CampID: camp.ID,
			Rooms: []model.Room{
				{Name: "New Room", Members: []model.User{user}},
			},
		}

		err := r.CreateRoomGroup(t.Context(), rgB)
		assert.ErrorIs(t, err, repository.ErrUserAlreadyAssigned)
	})

	t.Run("Success - User assigned in different camp", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		//　合宿Aを作成し、ユーザーを所属させておく
		campA := mustCreateCamp(t, r)
		groupA := mustCreateRoomGroup(t, r, campA.ID)
		user := mustCreateUser(t, r)
		_ = mustCreateRoom(t, r, groupA.ID, []model.User{user})

		//　別の合宿Bを作成
		campB := mustCreateCamp(t, r)

		// 合宿Aに所属済みのユーザーを、合宿Bの新しいRoomGroupに含める
		newRoomGroup := &model.RoomGroup{
			Name:   "Group in Camp B",
			CampID: campB.ID,
			Rooms: []model.Room{
				{
					Name: "Room in Camp B",
					Members: []model.User{
						user,
					},
				},
			},
		}

		// 重複エラーにならず成功することを期待
		err := r.CreateRoomGroup(t.Context(), newRoomGroup)

		// 検証
		assert.NoError(t, err)

		// DBに保存された結果を検証
		assert.NotZero(t, newRoomGroup.ID)
		assert.Len(t, newRoomGroup.Rooms[0].Members, 1)
		assert.Equal(t, user.ID, newRoomGroup.Rooms[0].Members[0].ID)
	})
}

func TestRepository_UpdateRoomGroup(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)

		newName := random.AlphaNumericString(t, 25)
		updatedRoomGroup := &model.RoomGroup{
			Name:   newName,
			CampID: camp.ID,
		}

		err := r.UpdateRoomGroup(t.Context(), roomGroup.ID, updatedRoomGroup)

		assert.NoError(t, err)

		// 更新が正しく反映されているか確認
		retrievedRoomGroup, err := r.GetRoomGroupByID(t.Context(), roomGroup.ID)

		require.NoError(t, err)
		assert.Equal(t, newName, retrievedRoomGroup.Name)
		assert.Equal(t, camp.ID, retrievedRoomGroup.CampID)
	})

	t.Run("Non-existent RoomGroup", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		nonExistentID := uint(random.PositiveInt(t))

		updatedRoomGroup := &model.RoomGroup{
			Name:   random.AlphaNumericString(t, 20),
			CampID: camp.ID,
		}

		err := r.UpdateRoomGroup(t.Context(), nonExistentID, updatedRoomGroup)

		assert.ErrorIs(t, err, repository.ErrRoomGroupNotFound)
	})

	t.Run("Zero CampID", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		newName := random.AlphaNumericString(t, 20)

		// CampIDは指定しない
		updatedRoomGroup := &model.RoomGroup{
			Name: newName,
		}

		err := r.UpdateRoomGroup(t.Context(), roomGroup.ID, updatedRoomGroup)

		assert.NoError(t, err)

		retrievedRoomGroup, err := r.GetRoomGroupByID(t.Context(), roomGroup.ID)

		assert.NoError(t, err)
		assert.Equal(t, updatedRoomGroup.Name, retrievedRoomGroup.Name)
		assert.NotZero(t, retrievedRoomGroup.ID)
		assert.Equal(t, newName, retrievedRoomGroup.Name)
		assert.Equal(t, camp.ID, retrievedRoomGroup.CampID) // CampIDは変更されないことを確認
	})

	t.Run("No Changes", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)

		updatedRoomGroup := &model.RoomGroup{
			Name:   roomGroup.Name,
			CampID: roomGroup.CampID,
		}

		err := r.UpdateRoomGroup(t.Context(), roomGroup.ID, updatedRoomGroup)

		assert.NoError(t, err)

		retrievedRoomGroup, err := r.GetRoomGroupByID(t.Context(), roomGroup.ID)

		require.NoError(t, err)
		assert.Equal(t, roomGroup.Name, retrievedRoomGroup.Name)
		assert.Equal(t, roomGroup.CampID, retrievedRoomGroup.CampID)
	})
}

func TestRepository_GetRoomGroupByID(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)

		retrievedRoomGroup, err := r.GetRoomGroupByID(t.Context(), roomGroup.ID)

		assert.NoError(t, err)

		if assert.NotNil(t, retrievedRoomGroup) {
			assert.Equal(t, roomGroup.ID, retrievedRoomGroup.ID)
			assert.Equal(t, roomGroup.Name, retrievedRoomGroup.Name)
			assert.Equal(t, roomGroup.CampID, retrievedRoomGroup.CampID)
			assert.NotNil(t, retrievedRoomGroup.Rooms) // Preloadされていることを確認
		}
	})

	t.Run("最新のステータスが含まれる", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})
		operator := mustCreateUser(t, r)

		mustSetRoomStatus(t, r, room.ID, &model.RoomStatus{
			Type:  "active",
			Topic: random.AlphaNumericString(t, 64),
		}, operator.ID)

		latestTopic := random.AlphaNumericString(t, 64)
		mustSetRoomStatus(t, r, room.ID, &model.RoomStatus{
			Type:  "inactive",
			Topic: latestTopic,
		}, operator.ID)

		retrievedRoomGroup, err := r.GetRoomGroupByID(t.Context(), roomGroup.ID)
		assert.NoError(t, err)

		if assert.NotNil(t, retrievedRoomGroup) &&
			assert.Len(t, retrievedRoomGroup.Rooms, 1) &&
			assert.NotNil(t, retrievedRoomGroup.Rooms[0].Status) {
			assert.Equal(t, "inactive", retrievedRoomGroup.Rooms[0].Status.Type)
			assert.Equal(t, latestTopic, retrievedRoomGroup.Rooms[0].Status.Topic)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		nonExistentID := uint(random.PositiveInt(t))

		retrievedRoomGroup, err := r.GetRoomGroupByID(t.Context(), nonExistentID)

		assert.ErrorIs(t, err, repository.ErrRoomGroupNotFound)
		assert.Nil(t, retrievedRoomGroup)
	})
}

func TestRepository_GetRoomGroups(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp1 := mustCreateCamp(t, r)
		camp2 := mustCreateCamp(t, r)
		// camp1に2つの部屋グループを作成
		roomGroup1 := mustCreateRoomGroup(t, r, camp1.ID)
		roomGroup2 := mustCreateRoomGroup(t, r, camp1.ID)
		// camp2に1つの部屋グループを作成（これは結果に含まれないはず）
		_ = mustCreateRoomGroup(t, r, camp2.ID)
		user1 := mustCreateUser(t, r)
		user2 := mustCreateUser(t, r)
		// room group1に部屋を作成
		room1 := &model.Room{
			Name:        random.AlphaNumericString(t, 10),
			RoomGroupID: roomGroup1.ID,
			Members:     []model.User{user1, user2},
		}
		err := r.CreateRoom(t.Context(), room1)

		require.NoError(t, err)

		// room group2に部屋を作成
		room2 := &model.Room{
			Name:        random.AlphaNumericString(t, 10),
			RoomGroupID: roomGroup2.ID,
			Members:     []model.User{user1},
		}
		err = r.CreateRoom(t.Context(), room2)

		require.NoError(t, err)

		roomGroups, err := r.GetRoomGroups(t.Context(), camp1.ID)

		assert.NoError(t, err)

		if assert.Len(t, roomGroups, 2) {
			assert.Equal(t, roomGroup1.ID, roomGroups[0].ID)
			assert.Equal(t, roomGroup1.Name, roomGroups[0].Name)

			if assert.Len(t, roomGroups[0].Rooms, 1) {
				assert.Equal(t, room1.ID, roomGroups[0].Rooms[0].ID)
				assert.Equal(t, room1.Name, roomGroups[0].Rooms[0].Name)
				assert.Equal(t, room1.RoomGroupID, roomGroups[0].Rooms[0].RoomGroupID)

				if assert.Len(t, roomGroups[0].Rooms[0].Members, 2) {
					expectedUserIDs := []string{user1.ID, user2.ID}
					actualUserIDs := []string{
						roomGroups[0].Rooms[0].Members[0].ID,
						roomGroups[0].Rooms[0].Members[1].ID,
					}

					assert.ElementsMatch(t, expectedUserIDs, actualUserIDs)
				}
			}

			assert.Equal(t, roomGroup2.ID, roomGroups[1].ID)
			assert.Equal(t, roomGroup2.Name, roomGroups[1].Name)

			if assert.Len(t, roomGroups[1].Rooms, 1) {
				assert.Equal(t, room2.ID, roomGroups[1].Rooms[0].ID)
				assert.Equal(t, room2.Name, roomGroups[1].Rooms[0].Name)
				assert.Equal(t, room2.RoomGroupID, roomGroups[1].Rooms[0].RoomGroupID)

				if assert.Len(t, roomGroups[1].Rooms[0].Members, 1) {
					assert.Equal(t, user1.ID, roomGroups[1].Rooms[0].Members[0].ID)
					assert.Equal(t, user1.IsStaff, roomGroups[1].Rooms[0].Members[0].IsStaff)
				}
			}
		}
	})

	t.Run("最新のステータスが含まれる", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		room := mustCreateRoom(t, r, roomGroup.ID, []model.User{})
		operator := mustCreateUser(t, r)

		mustSetRoomStatus(t, r, room.ID, &model.RoomStatus{
			Type:  "active",
			Topic: random.AlphaNumericString(t, 64),
		}, operator.ID)

		latestTopic := random.AlphaNumericString(t, 64)
		mustSetRoomStatus(t, r, room.ID, &model.RoomStatus{
			Type:  "inactive",
			Topic: latestTopic,
		}, operator.ID)

		roomGroups, err := r.GetRoomGroups(t.Context(), camp.ID)

		assert.NoError(t, err)

		if assert.Len(t, roomGroups, 1) &&
			assert.Len(t, roomGroups[0].Rooms, 1) &&
			assert.NotNil(t, roomGroups[0].Rooms[0].Status) {
			assert.Equal(t, "inactive", roomGroups[0].Rooms[0].Status.Type)
			assert.Equal(t, latestTopic, roomGroups[0].Rooms[0].Status.Topic)
		}
	})

	t.Run("Empty Result", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)

		roomGroups, err := r.GetRoomGroups(t.Context(), camp.ID)

		assert.NoError(t, err)
		assert.Empty(t, roomGroups)
	})

	t.Run("Non-existent Camp", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		nonExistentCampID := uint(random.PositiveInt(t))

		roomGroups, err := r.GetRoomGroups(t.Context(), nonExistentCampID)

		assert.ErrorIs(t, err, repository.ErrCampNotFound)
		assert.Nil(t, roomGroups)
	})
}

func TestRepository_DeleteRoomGroup(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		err := r.DeleteRoomGroup(t.Context(), roomGroup.ID)

		assert.NoError(t, err)

		// 削除されていることを確認
		_, err = r.GetRoomGroupByID(t.Context(), roomGroup.ID)
		assert.ErrorIs(t, err, repository.ErrRoomGroupNotFound)
	})

	t.Run("Non-existent RoomGroup", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		nonExistentID := uint(random.PositiveInt(t))
		err := r.DeleteRoomGroup(t.Context(), nonExistentID)

		assert.ErrorIs(t, err, repository.ErrRoomGroupNotFound)
	})

	t.Run("Delete RoomGroup with Rooms", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)
		user := mustCreateUser(t, r)
		// RoomGroupに関連するRoomを作成
		room := &model.Room{
			Name:        random.AlphaNumericString(t, 10),
			RoomGroupID: roomGroup.ID,
			Members:     []model.User{user},
		}
		err := r.CreateRoom(t.Context(), room)

		require.NoError(t, err)

		// RoomGroupを削除
		err = r.DeleteRoomGroup(t.Context(), roomGroup.ID)

		assert.NoError(t, err)

		_, err = r.GetRoomGroupByID(t.Context(), roomGroup.ID)

		assert.ErrorIs(t, err, repository.ErrRoomGroupNotFound)
	})
}

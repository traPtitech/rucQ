package gormrepository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

func mustCreateRoomGroup(t *testing.T, r *Repository, campID uint) *model.RoomGroup {
	t.Helper()

	roomGroup := &model.RoomGroup{
		Name:   random.AlphaNumericString(t, 20),
		CampID: campID,
	}

	err := r.CreateRoomGroup(context.Background(), roomGroup)
	if err != nil {
		t.Fatalf("failed to create room group: %v", err)
	}

	return roomGroup
}

func TestCreateRoomGroup(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := &model.RoomGroup{
			Name:   random.AlphaNumericString(t, 20),
			CampID: camp.ID,
		}

		err := r.CreateRoomGroup(context.Background(), roomGroup)

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

		err := r.CreateRoomGroup(context.Background(), roomGroup)

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

		err := r.CreateRoomGroup(context.Background(), roomGroup)

		// 外部キー制約違反によりエラーが発生することを確認
		assert.Error(t, err)
	})

	t.Run("Nil RoomGroup", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		err := r.CreateRoomGroup(context.Background(), nil)

		assert.Error(t, err)
	})
}

func TestUpdateRoomGroup(t *testing.T) {
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

		err := r.UpdateRoomGroup(context.Background(), roomGroup.ID, updatedRoomGroup)

		assert.NoError(t, err)

		// 更新が正しく反映されているか確認
		retrievedRoomGroup, err := r.GetRoomGroupByID(context.Background(), roomGroup.ID)
		assert.NoError(t, err)
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

		err := r.UpdateRoomGroup(context.Background(), nonExistentID, updatedRoomGroup)

		assert.ErrorIs(t, err, model.ErrNotFound)
	})

	t.Run("Nil RoomGroup", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)

		err := r.UpdateRoomGroup(context.Background(), roomGroup.ID, nil)

		assert.Error(t, err)
	})
}

func TestGetRoomGroupByID(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)

		retrievedRoomGroup, err := r.GetRoomGroupByID(context.Background(), roomGroup.ID)

		assert.NoError(t, err)
		assert.NotNil(t, retrievedRoomGroup)
		assert.Equal(t, roomGroup.ID, retrievedRoomGroup.ID)
		assert.Equal(t, roomGroup.Name, retrievedRoomGroup.Name)
		assert.Equal(t, roomGroup.CampID, retrievedRoomGroup.CampID)
		assert.NotNil(t, retrievedRoomGroup.Rooms) // Preloadされていることを確認
	})

	t.Run("Not Found", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		nonExistentID := uint(random.PositiveInt(t))

		retrievedRoomGroup, err := r.GetRoomGroupByID(context.Background(), nonExistentID)

		assert.Error(t, err)
		assert.Nil(t, retrievedRoomGroup)
	})
}

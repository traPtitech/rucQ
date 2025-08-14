package gormrepository

import (
	"testing"

	"github.com/stretchr/testify/assert"

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

	t.Run("Nil RoomGroup", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		err := r.CreateRoomGroup(t.Context(), nil)

		assert.Error(t, err)
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

		err := r.UpdateRoomGroup(t.Context(), nonExistentID, updatedRoomGroup)

		assert.ErrorIs(t, err, repository.ErrRoomGroupNotFound)
	})

	t.Run("Nil RoomGroup", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		roomGroup := mustCreateRoomGroup(t, r, camp.ID)

		err := r.UpdateRoomGroup(t.Context(), roomGroup.ID, nil)

		assert.Error(t, err)
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
		assert.NoError(t, err)
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

		retrievedRoomGroup, err := r.GetRoomGroupByID(t.Context(), nonExistentID)

		assert.ErrorIs(t, err, repository.ErrRoomGroupNotFound)
		assert.Nil(t, retrievedRoomGroup)
	})
}

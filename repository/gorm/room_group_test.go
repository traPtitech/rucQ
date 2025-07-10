package gorm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

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

package gormrepository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestRepository_CreateRollCall(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user1 := mustCreateUser(t, r)
		user2 := mustCreateUser(t, r)

		rollCall := model.RollCall{
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Options: []string{
				random.AlphaNumericString(t, 5),
				random.AlphaNumericString(t, 5),
				random.AlphaNumericString(t, 5),
			},
			Subjects: []model.User{user1, user2},
			CampID:   camp.ID,
		}

		err := r.CreateRollCall(t.Context(), &rollCall)

		assert.NoError(t, err)
		assert.NotZero(t, rollCall.ID)
		assert.Equal(t, camp.ID, rollCall.CampID)
	})

	t.Run("Camp not found", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		rollCall := model.RollCall{
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Options: []string{
				random.AlphaNumericString(t, 5),
				random.AlphaNumericString(t, 5),
			},
			Subjects: []model.User{},
			CampID:   uint(random.PositiveInt(t)), // 存在しないCampID
		}

		err := r.CreateRollCall(t.Context(), &rollCall)

		assert.ErrorIs(t, err, repository.ErrCampNotFound)
	})

	t.Run("Subject user not found", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)

		rollCall := model.RollCall{
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Options: []string{
				random.AlphaNumericString(t, 5),
				random.AlphaNumericString(t, 5),
			},
			Subjects: []model.User{{ID: random.AlphaNumericString(t, 32)}},
			CampID:   camp.ID,
		}

		err := r.CreateRollCall(t.Context(), &rollCall)

		assert.ErrorIs(t, err, repository.ErrUserNotFound)
	})
}

func TestRepository_GetRollCalls(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp1 := mustCreateCamp(t, r)
		camp2 := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)

		rollCall1 := mustCreateRollCall(t, r, camp1.ID, []model.User{user})
		rollCall2 := mustCreateRollCall(t, r, camp1.ID, []model.User{user})
		rollCall3 := mustCreateRollCall(t, r, camp2.ID, []model.User{user})

		rollCalls1, err := r.GetRollCalls(t.Context(), camp1.ID)

		assert.NoError(t, err)
		assert.Len(t, rollCalls1, 2)

		// IDで比較 (ロードされた関連データの違いを避けるため)
		rollCallIDs := []uint{rollCalls1[0].ID, rollCalls1[1].ID}
		assert.Contains(t, rollCallIDs, rollCall1.ID)
		assert.Contains(t, rollCallIDs, rollCall2.ID)

		rollCalls2, err := r.GetRollCalls(t.Context(), camp2.ID)

		assert.NoError(t, err)

		if assert.Len(t, rollCalls2, 1) {
			assert.Equal(t, rollCall3.ID, rollCalls2[0].ID)
		}
	})

	t.Run("Camp not found", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		_, err := r.GetRollCalls(t.Context(), uint(random.PositiveInt(t))) // 存在しないCampID

		assert.ErrorIs(t, err, repository.ErrCampNotFound)
	})

	t.Run("No roll calls", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)

		rollCalls, err := r.GetRollCalls(t.Context(), camp.ID)

		assert.NoError(t, err)
		assert.Empty(t, rollCalls)
	})
}

func mustCreateRollCall(
	t *testing.T,
	r *Repository,
	campID uint,
	subjects []model.User,
) model.RollCall {
	t.Helper()

	rollCall := model.RollCall{
		Name:        random.AlphaNumericString(t, 20),
		Description: random.AlphaNumericString(t, 100),
		Options: []string{
			random.AlphaNumericString(t, 5),
			random.AlphaNumericString(t, 5),
			random.AlphaNumericString(t, 5),
		},
		Subjects: subjects,
		CampID:   campID,
	}

	err := r.CreateRollCall(t.Context(), &rollCall)

	require.NoError(t, err)

	return rollCall
}

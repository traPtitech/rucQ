package gormrepository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestRepository_CreateRollCallReaction(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)
		rollCall := mustCreateRollCall(t, r, camp.ID, []model.User{user})

		reaction := model.RollCallReaction{
			Content:    random.AlphaNumericString(t, 10),
			UserID:     user.ID,
			RollCallID: rollCall.ID,
		}

		err := r.CreateRollCallReaction(t.Context(), &reaction)

		assert.NoError(t, err)
		assert.NotZero(t, reaction.ID)
		assert.Equal(t, rollCall.ID, reaction.RollCallID)
		assert.Equal(t, user.ID, reaction.UserID)
	})

	t.Run("RollCall not found", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		user := mustCreateUser(t, r)

		reaction := model.RollCallReaction{
			Content:    random.AlphaNumericString(t, 10),
			UserID:     user.ID,
			RollCallID: uint(random.PositiveInt(t)), // 存在しないRollCallID
		}

		err := r.CreateRollCallReaction(t.Context(), &reaction)

		assert.ErrorIs(t, err, repository.ErrRollCallNotFound)
	})
}

func TestRepository_GetRollCallReactions(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user1 := mustCreateUser(t, r)
		user2 := mustCreateUser(t, r)
		rollCall := mustCreateRollCall(t, r, camp.ID, []model.User{user1, user2})

		reaction1 := mustCreateRollCallReaction(t, r, rollCall.ID, user1.ID)
		reaction2 := mustCreateRollCallReaction(t, r, rollCall.ID, user2.ID)

		reactions, err := r.GetRollCallReactions(t.Context(), rollCall.ID)

		assert.NoError(t, err)

		if assert.Len(t, reactions, 2) {
			// IDでソートして比較
			if reactions[0].ID > reactions[1].ID {
				reactions[0], reactions[1] = reactions[1], reactions[0]
			}

			assert.Equal(t, reaction1.ID, reactions[0].ID)
			assert.Equal(t, reaction1.Content, reactions[0].Content)
			assert.Equal(t, reaction1.UserID, reactions[0].UserID)
			assert.Equal(t, reaction1.RollCallID, reactions[0].RollCallID)

			assert.Equal(t, reaction2.ID, reactions[1].ID)
			assert.Equal(t, reaction2.Content, reactions[1].Content)
			assert.Equal(t, reaction2.UserID, reactions[1].UserID)
			assert.Equal(t, reaction2.RollCallID, reactions[1].RollCallID)
		}
	})

	t.Run("Empty result", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)
		rollCall := mustCreateRollCall(t, r, camp.ID, []model.User{user})

		reactions, err := r.GetRollCallReactions(t.Context(), rollCall.ID)

		assert.NoError(t, err)
		assert.Empty(t, reactions)
	})

	t.Run("RollCall not found", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		_, err := r.GetRollCallReactions(t.Context(), uint(random.PositiveInt(t)))

		assert.ErrorIs(t, err, repository.ErrRollCallNotFound)
	})
}

func TestRepository_GetRollCallReactionByID(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)
		rollCall := mustCreateRollCall(t, r, camp.ID, []model.User{user})
		reaction := mustCreateRollCallReaction(t, r, rollCall.ID, user.ID)

		result, err := r.GetRollCallReactionByID(t.Context(), reaction.ID)

		assert.NoError(t, err)
		assert.Equal(t, reaction.ID, result.ID)
		assert.Equal(t, reaction.Content, result.Content)
		assert.Equal(t, reaction.UserID, result.UserID)
		assert.Equal(t, reaction.RollCallID, result.RollCallID)
	})

	t.Run("Not found", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		_, err := r.GetRollCallReactionByID(t.Context(), uint(random.PositiveInt(t)))

		assert.ErrorIs(t, err, repository.ErrRollCallReactionNotFound)
	})
}

func TestRepository_UpdateRollCallReaction(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)
		rollCall := mustCreateRollCall(t, r, camp.ID, []model.User{user})
		reaction := mustCreateRollCallReaction(t, r, rollCall.ID, user.ID)

		newContent := random.AlphaNumericString(t, 20)
		updateData := model.RollCallReaction{
			Content: newContent,
		}

		err := r.UpdateRollCallReaction(t.Context(), reaction.ID, &updateData)

		assert.NoError(t, err)

		// 更新されているかを確認
		updated, err := r.GetRollCallReactionByID(t.Context(), reaction.ID)
		assert.NoError(t, err)
		assert.Equal(t, newContent, updated.Content)
	})

	t.Run("Not found", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		updateData := model.RollCallReaction{
			Content: random.AlphaNumericString(t, 20),
		}

		err := r.UpdateRollCallReaction(t.Context(), uint(random.PositiveInt(t)), &updateData)

		assert.ErrorIs(t, err, repository.ErrRollCallReactionNotFound)
	})
}

func TestRepository_DeleteRollCallReaction(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)
		rollCall := mustCreateRollCall(t, r, camp.ID, []model.User{user})
		reaction := mustCreateRollCallReaction(t, r, rollCall.ID, user.ID)

		err := r.DeleteRollCallReaction(t.Context(), reaction.ID)

		assert.NoError(t, err)

		// 削除されているかを確認
		_, err = r.GetRollCallReactionByID(t.Context(), reaction.ID)
		assert.ErrorIs(t, err, repository.ErrRollCallReactionNotFound)
	})

	t.Run("Not found", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		err := r.DeleteRollCallReaction(t.Context(), uint(random.PositiveInt(t)))

		assert.ErrorIs(t, err, repository.ErrRollCallReactionNotFound)
	})
}

// mustCreateRollCallReaction creates a roll call reaction for testing purposes
func mustCreateRollCallReaction(
	t *testing.T,
	r *Repository,
	rollCallID uint,
	userID string,
) model.RollCallReaction {
	t.Helper()

	reaction := model.RollCallReaction{
		Content:    random.AlphaNumericString(t, 10),
		UserID:     userID,
		RollCallID: rollCallID,
	}

	err := r.CreateRollCallReaction(t.Context(), &reaction)

	require.NoError(t, err)

	return reaction
}

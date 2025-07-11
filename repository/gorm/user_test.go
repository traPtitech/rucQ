package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestGetOrCreateUser(t *testing.T) {
	t.Parallel()

	t.Run("Success (New User)", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		userID := random.AlphaNumericString(t, 32)
		user, err := r.GetOrCreateUser(t.Context(), userID)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.False(t, user.IsStaff)
	})

	t.Run("Success (Existing User)", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		user := mustCreateUser(t, r)
		got, err := r.GetOrCreateUser(t.Context(), user.ID)

		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, user.ID, got.ID)
		assert.False(t, got.IsStaff)
	})

	t.Run("Success (Concurrent Creation)", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		userID := random.AlphaNumericString(t, 32)

		concurrency := 10
		results := make(chan *model.User, concurrency)
		errs := make(chan error, concurrency)

		for range concurrency {
			go func() {
				user, err := r.GetOrCreateUser(t.Context(), userID)
				results <- user
				errs <- err
			}()
		}

		for range concurrency {
			user := <-results
			err := <-errs

			assert.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, userID, user.ID)
		}
	})
}

func TestUpdateUser(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		user := mustCreateUser(t, r)

		user.IsStaff = !user.IsStaff
		err := r.UpdateUser(t.Context(), &user)
		assert.NoError(t, err)

		// 更新が反映されているかを確認
		updatedUser, err := r.GetOrCreateUser(t.Context(), user.ID)
		assert.NoError(t, err)
		assert.Equal(t, user.IsStaff, updatedUser.IsStaff)
	})
}

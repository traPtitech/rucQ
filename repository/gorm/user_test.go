package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
}

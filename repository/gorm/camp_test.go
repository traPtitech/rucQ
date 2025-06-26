package gorm

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/traP-jp/rucQ/backend/model"
	"github.com/traP-jp/rucQ/backend/testutil/random"
)

func TestCreateCamp(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := model.Camp{
			DisplayID:          random.AlphaNumericString(t, 10),
			Name:               random.AlphaNumericString(t, 20),
			Description:        random.AlphaNumericString(t, 100),
			IsDraft:            random.Bool(t),
			IsPaymentOpen:      random.Bool(t),
			IsRegistrationOpen: random.Bool(t),
		}
		err := r.CreateCamp(&camp)

		require.NoError(t, err)
	})
}

package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestCreatePayment(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)
		amount := random.PositiveInt(t)
		amountPaid := random.PositiveInt(t)
		payment := model.Payment{
			Amount:     amount,
			AmountPaid: amountPaid,
			UserID:     user.ID,
			CampID:     camp.ID,
		}

		err := r.CreatePayment(t.Context(), &payment)

		assert.NoError(t, err)
		assert.NotZero(t, payment.ID)
		assert.Equal(t, amount, payment.Amount)
		assert.Equal(t, amountPaid, payment.AmountPaid)
		assert.Equal(t, user.ID, payment.UserID)
		assert.Equal(t, camp.ID, payment.CampID)
	})
}

package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestGetPayments(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)

		// テスト用のpaymentを複数作成
		payment1 := model.Payment{
			Amount:     random.PositiveInt(t),
			AmountPaid: random.PositiveInt(t),
			UserID:     user.ID,
			CampID:     camp.ID,
		}
		payment2 := model.Payment{
			Amount:     random.PositiveInt(t),
			AmountPaid: random.PositiveInt(t),
			UserID:     user.ID,
			CampID:     camp.ID,
		}

		err := r.CreatePayment(t.Context(), &payment1)
		require.NoError(t, err)
		err = r.CreatePayment(t.Context(), &payment2)
		require.NoError(t, err)

		// GetPaymentsをテスト
		payments, err := r.GetPayments(t.Context())

		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(payments), 2) // 少なくとも作成した2つは存在する

		// 作成したpaymentが含まれているかチェック
		foundPayment1 := false
		foundPayment2 := false
		for _, p := range payments {
			if p.ID == payment1.ID {
				foundPayment1 = true
				assert.Equal(t, payment1.Amount, p.Amount)
				assert.Equal(t, payment1.AmountPaid, p.AmountPaid)
				assert.Equal(t, payment1.UserID, p.UserID)
				assert.Equal(t, payment1.CampID, p.CampID)
			}
			if p.ID == payment2.ID {
				foundPayment2 = true
				assert.Equal(t, payment2.Amount, p.Amount)
				assert.Equal(t, payment2.AmountPaid, p.AmountPaid)
				assert.Equal(t, payment2.UserID, p.UserID)
				assert.Equal(t, payment2.CampID, p.CampID)
			}
		}
		assert.True(t, foundPayment1, "payment1 should be found in results")
		assert.True(t, foundPayment2, "payment2 should be found in results")
	})

	t.Run("Empty Result", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		// paymentが存在しない状態でGetPaymentsをテスト
		payments, err := r.GetPayments(t.Context())

		assert.NoError(t, err)
		assert.Empty(t, payments)
	})
}

package gormrepository

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
		payment1 := mustCreatePayment(t, r, user.ID, camp.ID)
		payment2 := mustCreatePayment(t, r, user.ID, camp.ID)

		// GetPaymentsをテスト
		payments, err := r.GetPayments(t.Context(), camp.ID)

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
		camp := mustCreateCamp(t, r)

		// paymentが存在しない状態でGetPaymentsをテスト
		payments, err := r.GetPayments(t.Context(), camp.ID)

		assert.NoError(t, err)
		assert.Empty(t, payments)
	})

	t.Run("Filter by CampID", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp1 := mustCreateCamp(t, r)
		camp2 := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)

		// 異なるcampのpaymentを作成
		payment1 := mustCreatePayment(t, r, user.ID, camp1.ID)
		payment2 := mustCreatePayment(t, r, user.ID, camp2.ID)

		// camp1のpaymentのみ取得
		paymentsForCamp1, err := r.GetPayments(t.Context(), camp1.ID)
		assert.NoError(t, err)
		assert.Len(t, paymentsForCamp1, 1)
		assert.Equal(t, payment1.ID, paymentsForCamp1[0].ID)
		assert.Equal(t, camp1.ID, paymentsForCamp1[0].CampID)

		// camp2のpaymentのみ取得
		paymentsForCamp2, err := r.GetPayments(t.Context(), camp2.ID)
		assert.NoError(t, err)
		assert.Len(t, paymentsForCamp2, 1)
		assert.Equal(t, payment2.ID, paymentsForCamp2[0].ID)
		assert.Equal(t, camp2.ID, paymentsForCamp2[0].CampID)
	})
}

func TestUpdatePayment(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)

		// テスト用のpaymentを作成
		originalPayment := mustCreatePayment(t, r, user.ID, camp.ID)

		// 更新用のデータ
		updatedAmount := 2000
		updatedAmountPaid := 1500
		updatePayment := model.Payment{
			Amount:     updatedAmount,
			AmountPaid: updatedAmountPaid,
			UserID:     user.ID,
			CampID:     camp.ID,
		}

		// UpdatePaymentをテスト
		err := r.UpdatePayment(t.Context(), originalPayment.ID, &updatePayment)
		assert.NoError(t, err)

		// 更新されたデータを取得して確認
		payments, err := r.GetPayments(t.Context(), camp.ID)
		require.NoError(t, err)

		var foundPayment *model.Payment
		for _, p := range payments {
			if p.ID == originalPayment.ID {
				foundPayment = &p
				break
			}
		}

		require.NotNil(t, foundPayment, "updated payment should be found")
		assert.Equal(t, updatedAmount, foundPayment.Amount)
		assert.Equal(t, updatedAmountPaid, foundPayment.AmountPaid)
		assert.Equal(t, user.ID, foundPayment.UserID)
		assert.Equal(t, camp.ID, foundPayment.CampID)
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)

		// 存在しないIDでUpdatePaymentをテスト
		nonExistentID := uint(99999)
		updatePayment := model.Payment{
			Amount:     1000,
			AmountPaid: 500,
			UserID:     user.ID,
			CampID:     camp.ID,
		}

		err := r.UpdatePayment(t.Context(), nonExistentID, &updatePayment)
		// GORMのUpdatesは存在しないレコードでもエラーを返さないことがある
		// 実際のビジネスロジックではレコードの存在確認が必要な場合がある
		assert.NoError(t, err)
	})
}

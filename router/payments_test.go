package router

import (
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestAdminPostPayment(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		req := api.AdminPostPaymentJSONRequestBody{
			Amount:     random.PositiveInt(t),
			AmountPaid: random.PositiveInt(t),
			UserId:     random.AlphaNumericString(t, 32),
		}
		adminUserID := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(&model.User{
				IsStaff: true,
			}, nil)
		h.repo.MockPaymentRepository.EXPECT().CreatePayment(gomock.Any(), gomock.Any()).Return(nil)

		res := h.expect.POST("/api/admin/camps/{campId}/payments", campID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().Status(http.StatusCreated).JSON().Object()

		res.Keys().ContainsOnly(
			"id", "amount", "amountPaid", "userId", "campId")
		res.Value("amount").Number().IsEqual(req.Amount)
		res.Value("amountPaid").Number().IsEqual(req.AmountPaid)
		res.Value("userId").String().IsEqual(req.UserId)
		res.Value("campId").Number().IsEqual(campID)
	})
}

func TestAdminGetPayments(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		adminUserID := random.AlphaNumericString(t, 32)

		// テスト用のpayment data
		payments := []model.Payment{
			{
				Model:      gorm.Model{ID: 1},
				Amount:     1000,
				AmountPaid: 500,
				UserID:     "user1",
				CampID:     uint(campID),
			},
			{
				Model:      gorm.Model{ID: 2},
				Amount:     2000,
				AmountPaid: 1500,
				UserID:     "user2",
				CampID:     uint(campID),
			},
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(&model.User{
				IsStaff: true,
			}, nil)
		h.repo.MockPaymentRepository.EXPECT().
			GetPayments(gomock.Any(), uint(campID)).
			Return(payments, nil)

		res := h.expect.GET("/api/admin/camps/{campId}/payments", campID).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().Status(http.StatusOK).JSON().Array()

		res.Length().IsEqual(2)

		// 最初のpaymentをチェック
		payment1 := res.Value(0).Object()
		payment1.Keys().ContainsOnly("id", "amount", "amountPaid", "userId", "campId")
		payment1.Value("id").Number().IsEqual(1)
		payment1.Value("amount").Number().IsEqual(1000)
		payment1.Value("amountPaid").Number().IsEqual(500)
		payment1.Value("userId").String().IsEqual("user1")
		payment1.Value("campId").Number().IsEqual(campID)

		// 2番目のpaymentをチェック
		payment2 := res.Value(1).Object()
		payment2.Keys().ContainsOnly("id", "amount", "amountPaid", "userId", "campId")
		payment2.Value("id").Number().IsEqual(2)
		payment2.Value("amount").Number().IsEqual(2000)
		payment2.Value("amountPaid").Number().IsEqual(1500)
		payment2.Value("userId").String().IsEqual("user2")
		payment2.Value("campId").Number().IsEqual(campID)
	})

	t.Run("Empty Result", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		adminUserID := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(&model.User{
				IsStaff: true,
			}, nil)
		h.repo.MockPaymentRepository.EXPECT().
			GetPayments(gomock.Any(), uint(campID)).
			Return([]model.Payment{}, nil)

		res := h.expect.GET("/api/admin/camps/{campId}/payments", campID).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().Status(http.StatusOK).JSON().Array()

		res.Length().IsEqual(0)
	})

	t.Run("Forbidden", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		userID := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&model.User{
				IsStaff: false, // 管理者ではないユーザー
			}, nil)

		h.expect.GET("/api/admin/camps/{campId}/payments", campID).
			WithHeader("X-Forwarded-User", userID).
			Expect().Status(http.StatusForbidden)
	})
}

func TestServer_AdminPutPayment(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		paymentID := random.PositiveInt(t)
		campID := random.PositiveInt(t)
		req := api.AdminPutPaymentJSONRequestBody{
			Amount:     2000,
			AmountPaid: 1500,
			UserId:     random.AlphaNumericString(t, 32),
		}
		adminUserID := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(&model.User{
				IsStaff: true,
			}, nil)
		h.repo.MockPaymentRepository.EXPECT().
			UpdatePayment(gomock.Any(), uint(paymentID), gomock.Any()).
			Return(nil)

		h.repo.MockPaymentRepository.EXPECT().
			GetPaymentByID(gomock.Any(), uint(paymentID)).
			Return(&model.Payment{
				Model:      gorm.Model{ID: uint(paymentID)},
				Amount:     req.Amount,
				AmountPaid: req.AmountPaid,
				UserID:     req.UserId,
				CampID:     uint(campID),
			}, nil)

		res := h.expect.PUT("/api/admin/payments/{paymentId}", paymentID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().Status(http.StatusOK).JSON().Object()

		res.Keys().ContainsOnly(
			"id", "amount", "amountPaid", "userId", "campId")
		res.Value("id").Number().IsEqual(paymentID)
		res.Value("amount").Number().IsEqual(req.Amount)
		res.Value("amountPaid").Number().IsEqual(req.AmountPaid)
		res.Value("userId").String().IsEqual(req.UserId)
		res.Value("campId").Number().IsEqual(campID)
	})

	t.Run("Forbidden", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		paymentID := random.PositiveInt(t)
		req := api.AdminPutPaymentJSONRequestBody{
			Amount:     2000,
			AmountPaid: 1500,
			UserId:     random.AlphaNumericString(t, 32),
		}
		userID := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&model.User{
				IsStaff: false, // 管理者ではないユーザー
			}, nil)

		h.expect.PUT("/api/admin/payments/{paymentId}", paymentID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", userID).
			Expect().Status(http.StatusForbidden)
	})

	t.Run("Payment Not Found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		paymentID := random.PositiveInt(t)
		req := api.AdminPutPaymentJSONRequestBody{
			Amount:     random.PositiveInt(t),
			AmountPaid: random.PositiveInt(t),
			UserId:     random.AlphaNumericString(t, 32),
		}
		adminUserID := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(&model.User{
				IsStaff: true,
			}, nil).Times(1)
		h.repo.MockPaymentRepository.EXPECT().
			UpdatePayment(gomock.Any(), uint(paymentID), gomock.Any()).
			Return(repository.ErrPaymentNotFound).Times(1)

		h.expect.PUT("/api/admin/payments/{paymentId}", paymentID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().Status(http.StatusNotFound)
	})

	t.Run("Bad Request", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		paymentID := random.PositiveInt(t)
		adminUserID := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(&model.User{
				IsStaff: true,
			}, nil)

		// 不正なJSONリクエスト
		h.expect.PUT("/api/admin/payments/{paymentId}", paymentID).
			WithHeader("Content-Type", "application/json").
			WithText("invalid json").
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().Status(http.StatusBadRequest)
	})
}

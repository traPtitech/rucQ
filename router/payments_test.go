package router

import (
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
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
			CampId:     campID,
		}
		adminUserID := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().GetOrCreateUser(gomock.Any(), adminUserID).Return(&model.User{
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

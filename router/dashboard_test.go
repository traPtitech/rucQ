package router

import (
	"errors"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestServer_GetDashboard(t *testing.T) {
	t.Parallel()

	t.Run("Success - Without Payment and Room", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		userID := random.AlphaNumericString(t, 32)

		// モックの設定: ユーザーが参加者である場合
		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), uint(campID), userID).
			Return(true, nil).
			Times(1)
		h.repo.MockPaymentRepository.EXPECT().
			GetPaymentByUserID(gomock.Any(), uint(campID), userID).
			Return(nil, repository.ErrPaymentNotFound).
			Times(1)
		h.repo.MockRoomRepository.EXPECT().
			GetRoomByUserID(gomock.Any(), uint(campID), userID).
			Return(nil, repository.ErrRoomNotFound).
			Times(1)

		res := h.expect.GET("/api/camps/{campId}/me", campID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		res.Keys().ContainsOnly("id")
		res.Value("id").String().IsEqual(userID)
	})

	t.Run("Success - With Payment and Room", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		userID := random.AlphaNumericString(t, 32)

		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), uint(campID), userID).
			Return(true, nil).
			Times(1)
		payment := &model.Payment{
			Model: gorm.Model{
				ID: uint(random.PositiveInt(t)),
			},
			Amount:     random.PositiveInt(t),
			AmountPaid: random.PositiveInt(t),
			UserID:     userID,
			CampID:     uint(campID),
		}
		h.repo.MockPaymentRepository.EXPECT().
			GetPaymentByUserID(gomock.Any(), uint(campID), userID).
			Return(payment, nil).
			Times(1)

		room := &model.Room{
			Model: gorm.Model{
				ID: uint(random.PositiveInt(t)),
			},
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupID: uint(random.PositiveInt(t)),
			Members: []model.User{
				{
					ID:      userID,
					IsStaff: true,
				},
			},
		}

		h.repo.MockRoomRepository.EXPECT().
			GetRoomByUserID(gomock.Any(), uint(campID), userID).
			Return(room, nil).
			Times(1)

		res := h.expect.GET("/api/camps/{campId}/me", campID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		res.Keys().ContainsOnly("id", "payment", "room")
		res.Value("id").String().IsEqual(userID)

		paymentRes := res.Value("payment").Object()

		paymentRes.Keys().ContainsOnly("id", "amount", "amountPaid", "campId", "userId")
		paymentRes.Value("id").Number().IsEqual(payment.ID)
		paymentRes.Value("amount").Number().IsEqual(payment.Amount)
		paymentRes.Value("amountPaid").Number().IsEqual(payment.AmountPaid)
		paymentRes.Value("campId").Number().IsEqual(payment.CampID)
		paymentRes.Value("userId").String().IsEqual(payment.UserID)

		roomRes := res.Value("room").Object()

		roomRes.Keys().ContainsOnly("id", "name", "members", "status")
		roomRes.Value("id").Number().IsEqual(room.ID)
		roomRes.Value("name").String().IsEqual(room.Name)
		roomRes.Value("status").Object().
			HasValue("topic", "").
			Value("type").IsNull()

		members := roomRes.Value("members").Array()

		members.Length().IsEqual(1)

		member := members.Value(0).Object()

		member.Keys().ContainsOnly("id", "isStaff")
		member.Value("id").String().IsEqual(userID)
		member.Value("isStaff").Boolean().IsTrue()
	})

	t.Run("Success - With Payment only", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		userID := random.AlphaNumericString(t, 32)

		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), uint(campID), userID).
			Return(true, nil).
			Times(1)

		payment := &model.Payment{
			Model: gorm.Model{
				ID: uint(random.PositiveInt(t)),
			},
			Amount:     random.PositiveInt(t),
			AmountPaid: random.PositiveInt(t),
			UserID:     userID,
			CampID:     uint(campID),
		}

		h.repo.MockPaymentRepository.EXPECT().
			GetPaymentByUserID(gomock.Any(), uint(campID), userID).
			Return(payment, nil).
			Times(1)
		h.repo.MockRoomRepository.EXPECT().
			GetRoomByUserID(gomock.Any(), uint(campID), userID).
			Return(nil, repository.ErrRoomNotFound).
			Times(1)

		res := h.expect.GET("/api/camps/{campId}/me", campID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		res.Keys().ContainsOnly("id", "payment")
		res.Value("id").String().IsEqual(userID)

		paymentRes := res.Value("payment").Object()

		paymentRes.Keys().ContainsOnly("id", "amount", "amountPaid", "campId", "userId")
		paymentRes.Value("id").Number().IsEqual(payment.ID)
		paymentRes.Value("amount").Number().IsEqual(payment.Amount)
		paymentRes.Value("amountPaid").Number().IsEqual(payment.AmountPaid)
		paymentRes.Value("campId").Number().IsEqual(payment.CampID)
		paymentRes.Value("userId").String().IsEqual(payment.UserID)
	})

	t.Run("Success - With Room only", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		userID := random.AlphaNumericString(t, 32)

		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), uint(campID), userID).
			Return(true, nil).
			Times(1)
		h.repo.MockPaymentRepository.EXPECT().
			GetPaymentByUserID(gomock.Any(), uint(campID), userID).
			Return(nil, repository.ErrPaymentNotFound).
			Times(1)
		room := &model.Room{
			Model: gorm.Model{
				ID: uint(random.PositiveInt(t)),
			},
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupID: uint(random.PositiveInt(t)),
			Members: []model.User{
				{
					ID:      userID,
					IsStaff: false,
				},
			},
		}

		h.repo.MockRoomRepository.EXPECT().
			GetRoomByUserID(gomock.Any(), uint(campID), userID).
			Return(room, nil).
			Times(1)

		res := h.expect.GET("/api/camps/{campId}/me", campID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		res.Keys().ContainsOnly("id", "room")
		res.Value("id").String().IsEqual(userID)

		roomRes := res.Value("room").Object()

		roomRes.Keys().ContainsOnly("id", "name", "members", "status")
		roomRes.Value("id").Number().IsEqual(room.ID)
		roomRes.Value("name").String().IsEqual(room.Name)
		roomRes.Value("status").Object().
			HasValue("topic", "").
			Value("type").IsNull()

		members := roomRes.Value("members").Array()
		members.Length().IsEqual(1)

		member := members.Value(0).Object()
		member.Keys().ContainsOnly("id", "isStaff")
		member.Value("id").String().IsEqual(userID)
		member.Value("isStaff").Boolean().IsFalse()
	})

	t.Run("Not Found - User is not a participant", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		userID := random.AlphaNumericString(t, 32)

		// モックの設定: ユーザーが参加者でない場合
		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), uint(campID), userID).
			Return(false, nil)

		h.expect.GET("/api/camps/{campId}/me", campID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			HasValue("message", "User is not a participant of this camp")
	})

	t.Run("Not Found - Camp does not exist", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		userID := random.AlphaNumericString(t, 32)

		// モックの設定: キャンプが存在しない場合
		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), uint(campID), userID).
			Return(false, model.ErrNotFound)

		h.expect.GET("/api/camps/{campId}/me", campID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			HasValue("message", "Camp not found")
	})

	t.Run("Internal Server Error - Camp repo error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		userID := random.AlphaNumericString(t, 32)

		// モックの設定: リポジトリでエラーが発生する場合
		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), uint(campID), userID).
			Return(false, errors.New("database error"))

		h.expect.GET("/api/camps/{campId}/me", campID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("Internal Server Error - Payment repo error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		userID := random.AlphaNumericString(t, 32)

		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), uint(campID), userID).
			Return(true, nil).
			Times(1)
		h.repo.MockPaymentRepository.EXPECT().
			GetPaymentByUserID(gomock.Any(), uint(campID), userID).
			Return(nil, errors.New("database error")).
			Times(1)
		h.expect.GET("/api/camps/{campId}/me", campID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("Internal Server Error - Room repo error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		userID := random.AlphaNumericString(t, 32)

		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), uint(campID), userID).
			Return(true, nil).
			Times(1)

		h.repo.MockPaymentRepository.EXPECT().
			GetPaymentByUserID(gomock.Any(), uint(campID), userID).
			Return(nil, repository.ErrPaymentNotFound).
			Times(1)
		h.repo.MockRoomRepository.EXPECT().
			GetRoomByUserID(gomock.Any(), uint(campID), userID).
			Return(nil, errors.New("database error")).
			Times(1)
		h.expect.GET("/api/camps/{campId}/me", campID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusInternalServerError)
	})
}

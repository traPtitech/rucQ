package router

import (
	"net/http"
	"testing"
	"time"

	"github.com/oapi-codegen/runtime/types"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestGetCamps(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		dateStart := random.Time(t)
		dateEnd := dateStart.Add(time.Duration(random.PositiveInt(t)))
		camp := model.Camp{
			Model: gorm.Model{
				ID: uint(random.PositiveInt(t)),
			},
			DisplayID:          random.AlphaNumericString(t, 10),
			Name:               random.AlphaNumericString(t, 20),
			Guidebook:          random.AlphaNumericString(t, 100),
			IsDraft:            random.Bool(t),
			IsPaymentOpen:      random.Bool(t),
			IsRegistrationOpen: random.Bool(t),
			DateStart:          dateStart,
			DateEnd:            dateEnd,
		}

		h.repo.MockCampRepository.EXPECT().GetCamps().Return([]model.Camp{camp}, nil)

		res := h.expect.GET("/api/camps").Expect().Status(http.StatusOK).JSON().Array()

		res.Length().IsEqual(1)

		val := res.Value(0).Object()

		val.Keys().ContainsOnly(
			"id", "displayId", "name", "description", "isDraft", "isPaymentOpen",
			"isRegistrationOpen", "dateStart", "dateEnd")
		val.Value("id").Number().IsEqual(camp.ID)
		val.Value("displayId").String().IsEqual(camp.DisplayID)
		val.Value("name").String().IsEqual(camp.Name)
		val.Value("description").String().IsEqual(camp.Guidebook)
		val.Value("isDraft").Boolean().IsEqual(camp.IsDraft)
		val.Value("isPaymentOpen").Boolean().IsEqual(camp.IsPaymentOpen)
		val.Value("isRegistrationOpen").Boolean().IsEqual(camp.IsRegistrationOpen)
		val.Value("dateStart").String().IsEqual(camp.DateStart.Format(time.DateOnly))
		val.Value("dateEnd").String().IsEqual(camp.DateEnd.Format(time.DateOnly))
	})
}

func TestAdminPostCamp(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		dateStart := random.Time(t)
		dateEnd := dateStart.Add(time.Duration(random.PositiveInt(t)))
		req := api.AdminPostCampJSONRequestBody{
			DisplayId:          random.AlphaNumericString(t, 10),
			Name:               random.AlphaNumericString(t, 20),
			Guidebook:          random.AlphaNumericString(t, 100),
			IsDraft:            random.Bool(t),
			IsPaymentOpen:      random.Bool(t),
			IsRegistrationOpen: random.Bool(t),
			DateStart:          types.Date{Time: dateStart},
			DateEnd:            types.Date{Time: dateEnd},
		}
		username := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().GetOrCreateUser(gomock.Any(), username).Return(&model.User{IsStaff: true}, nil)
		h.repo.MockCampRepository.EXPECT().CreateCamp(gomock.Any()).Return(nil)

		res := h.expect.POST("/api/admin/camps").
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusCreated).JSON().Object()

		res.Keys().ContainsOnly(
			"id", "displayId", "name", "guidebook", "isDraft", "isPaymentOpen",
			"isRegistrationOpen", "dateStart", "dateEnd")
		res.Value("displayId").String().IsEqual(req.DisplayId)
		res.Value("name").String().IsEqual(req.Name)
		res.Value("guidebook").String().IsEqual(req.Guidebook)
		res.Value("isDraft").Boolean().IsEqual(req.IsDraft)
		res.Value("isPaymentOpen").Boolean().IsEqual(req.IsPaymentOpen)
		res.Value("isRegistrationOpen").Boolean().IsEqual(req.IsRegistrationOpen)
		res.Value("dateStart").String().IsEqual(req.DateStart.Format(time.DateOnly))
		res.Value("dateEnd").String().IsEqual(req.DateEnd.Format(time.DateOnly))
	})
}

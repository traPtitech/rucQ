package router

import (
	"net/http"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/traP-jp/rucQ/backend/model"
	"github.com/traP-jp/rucQ/backend/testutil/random"
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
			Description:        random.AlphaNumericString(t, 100),
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
		val.Value("description").String().IsEqual(camp.Description)
		val.Value("isDraft").Boolean().IsEqual(camp.IsDraft)
		val.Value("isPaymentOpen").Boolean().IsEqual(camp.IsPaymentOpen)
		val.Value("isRegistrationOpen").Boolean().IsEqual(camp.IsRegistrationOpen)
		val.Value("dateStart").String().IsEqual(camp.DateStart.Format(time.DateOnly))
		val.Value("dateEnd").String().IsEqual(camp.DateEnd.Format(time.DateOnly))
	})
}

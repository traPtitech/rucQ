package router

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestGetEvents(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))
		timeStart1 := random.Time(t)
		timeEnd1 := timeStart1.Add(time.Duration(random.PositiveInt(t)))
		userID := random.AlphaNumericString(t, 32)
		color := "blue" // TODO: validなものからランダムに選ぶ

		durationEvent := model.Event{
			Model: gorm.Model{
				ID: uint(random.PositiveInt(t)),
			},
			Type:         model.EventTypeDuration,
			Name:         random.AlphaNumericString(t, 20),
			Description:  random.AlphaNumericString(t, 100),
			Location:     random.AlphaNumericString(t, 50),
			TimeStart:    timeStart1,
			TimeEnd:      &timeEnd1,
			OrganizerID:  &userID,
			DisplayColor: &color,
		}

		timeStart2 := random.Time(t)

		momentEvent := model.Event{
			Model: gorm.Model{
				ID: uint(random.PositiveInt(t)),
			},
			Type:        model.EventTypeMoment,
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Location:    random.AlphaNumericString(t, 50),
			TimeStart:   timeStart2,
		}

		timeStart3 := random.Time(t)
		timeEnd3 := timeStart3.Add(time.Duration(random.PositiveInt(t)))

		officialEvent := model.Event{
			Model: gorm.Model{
				ID: uint(random.PositiveInt(t)),
			},
			Type:        model.EventTypeOfficial,
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Location:    random.AlphaNumericString(t, 50),
			TimeStart:   timeStart3,
			TimeEnd:     &timeEnd3,
		}

		h.repo.MockEventRepository.EXPECT().GetEvents().Return([]model.Event{
			durationEvent,
			momentEvent,
			officialEvent,
		}, nil)

		res := h.expect.GET(fmt.Sprintf("/api/camps/%d/events", campID)).Expect().Status(http.StatusOK).JSON().Array()

		res.Length().IsEqual(3)

		res1 := res.Value(0).Object()

		res1.Keys().ContainsOnly("id", "type", "name", "description", "location", "timeStart",
			"timeEnd", "organizerId", "displayColor")
		res1.Value("id").Number().IsEqual(durationEvent.ID)
		res1.Value("type").String().IsEqual(string(model.EventTypeDuration))
		res1.Value("name").String().IsEqual(durationEvent.Name)
		res1.Value("description").String().IsEqual(durationEvent.Description)
		res1.Value("location").String().IsEqual(durationEvent.Location)
		res1.Value("timeStart").String().AsDateTime().IsEqual(durationEvent.TimeStart)
		res1.Value("timeEnd").String().AsDateTime().IsEqual(*durationEvent.TimeEnd)
		res1.Value("organizerId").String().IsEqual(*durationEvent.OrganizerID)
		res1.Value("displayColor").String().IsEqual(*durationEvent.DisplayColor)

		res2 := res.Value(1).Object()

		res2.Keys().ContainsOnly("id", "type", "name", "description", "location", "time")
		res2.Value("id").Number().IsEqual(momentEvent.ID)
		res2.Value("type").String().IsEqual(string(model.EventTypeMoment))
		res2.Value("name").String().IsEqual(momentEvent.Name)
		res2.Value("description").String().IsEqual(momentEvent.Description)
		res2.Value("location").String().IsEqual(momentEvent.Location)
		res2.Value("time").String().AsDateTime().IsEqual(momentEvent.TimeStart)

		res3 := res.Value(2).Object()
		res3.Keys().ContainsOnly("id", "type", "name", "description", "location", "timeStart", "timeEnd")
		res3.Value("id").Number().IsEqual(officialEvent.ID)
		res3.Value("type").String().IsEqual(string(model.EventTypeOfficial))
		res3.Value("name").String().IsEqual(officialEvent.Name)
		res3.Value("description").String().IsEqual(officialEvent.Description)
		res3.Value("location").String().IsEqual(officialEvent.Location)
		res3.Value("timeStart").String().AsDateTime().IsEqual(officialEvent.TimeStart)
		res3.Value("timeEnd").String().AsDateTime().IsEqual(*officialEvent.TimeEnd)
	})
}

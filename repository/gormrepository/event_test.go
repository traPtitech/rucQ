package gormrepository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestGetEvents(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp1 := mustCreateCamp(t, r)
		camp2 := mustCreateCamp(t, r)
		event1 := mustCreateEvent(t, r, camp1.ID)
		event2 := mustCreateEvent(t, r, camp1.ID)
		event3 := mustCreateEvent(t, r, camp2.ID)

		events1, err := r.GetEvents(t.Context(), camp1.ID)

		assert.NoError(t, err)
		assert.Len(t, events1, 2)
		assert.Contains(t, events1, event1)
		assert.Contains(t, events1, event2)

		events2, err := r.GetEvents(t.Context(), camp2.ID)

		assert.NoError(t, err)
		assert.Len(t, events2, 1)
		assert.Contains(t, events2, event3)
	})
}

func TestCreateEvent(t *testing.T) {
	t.Parallel()

	t.Run("Success (Duration Event)", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		name := random.AlphaNumericString(t, 20)
		description := random.AlphaNumericString(t, 100)
		location := random.AlphaNumericString(t, 50)
		timeStart := random.Time(t)
		timeEnd := timeStart.Add(time.Duration(random.PositiveInt(t)))
		color := random.AlphaNumericString(t, 10)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)
		event := model.Event{
			Type:         model.EventTypeDuration,
			Name:         name,
			Description:  description,
			Location:     location,
			TimeStart:    timeStart,
			TimeEnd:      &timeEnd,
			OrganizerID:  &user.ID,
			DisplayColor: &color,
			CampID:       camp.ID,
		}

		err := r.CreateEvent(&event)

		assert.NoError(t, err)
		assert.NotZero(t, event.ID)
		assert.Equal(t, model.EventTypeDuration, event.Type)
		assert.Equal(t, camp.ID, event.CampID)
		assert.Equal(t, name, event.Name)
		assert.Equal(t, description, event.Description)
		assert.Equal(t, location, event.Location)
		assert.Equal(t, timeStart, event.TimeStart)
		assert.Equal(t, timeEnd, *event.TimeEnd)
		assert.Equal(t, user.ID, *event.OrganizerID)
		assert.Equal(t, color, *event.DisplayColor)
	})

	t.Run("Success (Moment Event)", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		name := random.AlphaNumericString(t, 20)
		description := random.AlphaNumericString(t, 100)
		location := random.AlphaNumericString(t, 50)
		time := random.Time(t)
		camp := mustCreateCamp(t, r)
		event := model.Event{
			Type:        model.EventTypeMoment,
			Name:        name,
			Description: description,
			Location:    location,
			TimeStart:   time,
			CampID:      camp.ID,
		}

		err := r.CreateEvent(&event)

		assert.NoError(t, err)
		assert.NotZero(t, event.ID)
		assert.Equal(t, model.EventTypeMoment, event.Type)
		assert.Equal(t, camp.ID, event.CampID)
		assert.Equal(t, name, event.Name)
		assert.Equal(t, description, event.Description)
		assert.Equal(t, location, event.Location)
		assert.Equal(t, time, event.TimeStart)
	})

	t.Run("Success (Official Event)", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		name := random.AlphaNumericString(t, 20)
		description := random.AlphaNumericString(t, 100)
		location := random.AlphaNumericString(t, 50)
		timeStart := random.Time(t)
		timeEnd := timeStart.Add(time.Duration(random.PositiveInt(t)))
		camp := mustCreateCamp(t, r)
		event := model.Event{
			Type:        model.EventTypeOfficial,
			Name:        name,
			Description: description,
			Location:    location,
			TimeStart:   timeStart,
			TimeEnd:     &timeEnd,
			CampID:      camp.ID,
		}

		err := r.CreateEvent(&event)

		assert.NoError(t, err)
		assert.NotZero(t, event.ID)
		assert.Equal(t, model.EventTypeOfficial, event.Type)
		assert.Equal(t, camp.ID, event.CampID)
		assert.Equal(t, name, event.Name)
		assert.Equal(t, description, event.Description)
		assert.Equal(t, location, event.Location)
		assert.Equal(t, timeStart, event.TimeStart)
		assert.Equal(t, timeEnd, *event.TimeEnd)
	})

	t.Run("Success (Multiple Events)", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)
		event1 := model.Event{
			Type:         model.EventTypeMoment,
			Name:         random.AlphaNumericString(t, 20),
			Description:  random.AlphaNumericString(t, 100),
			Location:     random.AlphaNumericString(t, 50),
			TimeStart:    random.Time(t),
			CampID:       camp.ID,
			OrganizerID:  &user.ID,
			DisplayColor: nil,
		}
		event2 := model.Event{
			Type:         model.EventTypeMoment,
			Name:         random.AlphaNumericString(t, 20),
			Description:  random.AlphaNumericString(t, 100),
			Location:     random.AlphaNumericString(t, 50),
			TimeStart:    random.Time(t),
			CampID:       camp.ID,
			OrganizerID:  &user.ID,
			DisplayColor: nil,
		}

		err := r.CreateEvent(&event1)

		assert.NoError(t, err)
		assert.NotZero(t, event1.ID)

		err = r.CreateEvent(&event2)

		assert.NoError(t, err)
		assert.NotZero(t, event2.ID)
	})

	t.Run("Failure (Organizer Not Found)", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		userID := random.AlphaNumericString(t, 32) // 存在しないユーザーID
		event := model.Event{
			Type:        model.EventTypeMoment,
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Location:    random.AlphaNumericString(t, 50),
			TimeStart:   random.Time(t),
			CampID:      camp.ID,
			OrganizerID: &userID,
		}
		err := r.CreateEvent(&event)

		assert.Error(t, err)
		// TODO: 具体的なエラー内容を確認するためのアサーションを追加する
	})
}

package router

import (
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/stretchr/testify/require"

	"github.com/traPtitech/rucQ/api"
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
		color := string(random.SelectFrom(
			t,
			api.DurationEventRequestDisplayColorBlue,
			api.DurationEventRequestDisplayColorGreen,
			api.DurationEventRequestDisplayColorOrange,
			api.DurationEventRequestDisplayColorPink,
			api.DurationEventRequestDisplayColorPurple,
			api.DurationEventRequestDisplayColorRed,
		))

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

		h.repo.MockEventRepository.EXPECT().GetEvents(gomock.Any(), campID).Return([]model.Event{
			durationEvent,
			momentEvent,
			officialEvent,
		}, nil)

		res := h.expect.GET("/api/camps/{campId}/events", campID).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

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
		res3.Keys().
			ContainsOnly("id", "type", "name", "description", "location", "timeStart", "timeEnd")
		res3.Value("id").Number().IsEqual(officialEvent.ID)
		res3.Value("type").String().IsEqual(string(model.EventTypeOfficial))
		res3.Value("name").String().IsEqual(officialEvent.Name)
		res3.Value("description").String().IsEqual(officialEvent.Description)
		res3.Value("location").String().IsEqual(officialEvent.Location)
		res3.Value("timeStart").String().AsDateTime().IsEqual(officialEvent.TimeStart)
		res3.Value("timeEnd").String().AsDateTime().IsEqual(*officialEvent.TimeEnd)
	})
}

func TestServer_PostEvent(t *testing.T) {
	t.Parallel()

	t.Run("Success - Duration event with organizer who is camp participant", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		organizerID := random.AlphaNumericString(t, 32)
		timeStart := random.Time(t)
		timeEnd := timeStart.Add(time.Duration(random.PositiveInt(t)))
		displayColor := random.SelectFrom(
			t,
			api.DurationEventRequestDisplayColorBlue,
			api.DurationEventRequestDisplayColorGreen,
			api.DurationEventRequestDisplayColorOrange,
			api.DurationEventRequestDisplayColorPink,
			api.DurationEventRequestDisplayColorPurple,
			api.DurationEventRequestDisplayColorRed,
		)
		reqBody := api.DurationEventRequest{
			Type:         api.DurationEventRequestTypeDuration,
			Name:         random.AlphaNumericString(t, 20),
			Description:  random.AlphaNumericString(t, 100),
			Location:     random.AlphaNumericString(t, 50),
			TimeStart:    timeStart,
			TimeEnd:      timeEnd,
			OrganizerId:  organizerID,
			DisplayColor: displayColor,
		}
		user := model.User{
			ID:      userID,
			IsStaff: false,
		}
		createdEvent := model.Event{
			Model:        gorm.Model{ID: uint(random.PositiveInt(t))},
			Type:         model.EventTypeDuration,
			Name:         reqBody.Name,
			Description:  reqBody.Description,
			Location:     reqBody.Location,
			TimeStart:    reqBody.TimeStart,
			TimeEnd:      &reqBody.TimeEnd,
			OrganizerID:  &reqBody.OrganizerId,
			DisplayColor: (*string)(&reqBody.DisplayColor),
			CampID:       campID,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&user, nil)
		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), campID, organizerID).
			Return(true, nil)
		h.repo.MockEventRepository.EXPECT().
			CreateEvent(gomock.Any()).
			DoAndReturn(func(event *model.Event) error {
				event.ID = createdEvent.ID
				return nil
			})

		var eventRequest api.EventRequest

		err := eventRequest.FromDurationEventRequest(reqBody)

		require.NoError(t, err)

		res := h.expect.POST("/api/camps/{campId}/events", campID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(eventRequest).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object()

		res.Keys().
			ContainsOnly("id", "type", "name", "description", "location", "timeStart", "timeEnd", "organizerId", "displayColor")
		res.Value("id").Number().IsEqual(createdEvent.ID)
		res.Value("type").String().IsEqual(string(model.EventTypeDuration))
		res.Value("name").String().IsEqual(reqBody.Name)
		res.Value("description").String().IsEqual(reqBody.Description)
		res.Value("location").String().IsEqual(reqBody.Location)
		res.Value("timeStart").String().AsDateTime().IsEqual(reqBody.TimeStart)
		res.Value("timeEnd").String().AsDateTime().IsEqual(reqBody.TimeEnd)
		res.Value("organizerId").String().IsEqual(reqBody.OrganizerId)
		res.Value("displayColor").String().IsEqual(string(reqBody.DisplayColor))
	})

	t.Run("Failure - Organizer is not camp participant", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		organizerID := random.AlphaNumericString(t, 32)
		timeStart := random.Time(t)
		timeEnd := timeStart.Add(time.Duration(random.PositiveInt(t)))
		reqBody := api.DurationEventRequest{
			Type:        api.DurationEventRequestTypeDuration,
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Location:    random.AlphaNumericString(t, 50),
			TimeStart:   timeStart,
			TimeEnd:     timeEnd,
			OrganizerId: organizerID,
		}
		user := model.User{
			ID:      userID,
			IsStaff: false,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&user, nil)

		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), campID, organizerID).
			Return(false, nil)

		var eventRequest api.EventRequest

		err := eventRequest.FromDurationEventRequest(reqBody)

		require.NoError(t, err)

		h.expect.POST("/api/camps/{campId}/events", campID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(eventRequest).
			Expect().
			Status(http.StatusBadRequest).
			JSON().
			Object().
			Value("message").String().Contains("Organizer must be a participant of the camp")
	})

	t.Run("Success - Official event by staff user without organizer", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		timeStart := random.Time(t)
		timeEnd := timeStart.Add(time.Duration(random.PositiveInt(t)))
		reqBody := api.OfficialEventRequest{
			Type:        api.OfficialEventRequestTypeOfficial,
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Location:    random.AlphaNumericString(t, 50),
			TimeStart:   timeStart,
			TimeEnd:     timeEnd,
		}
		user := model.User{
			ID:      userID,
			IsStaff: true, // Staff user
		}
		createdEvent := model.Event{
			Model:       gorm.Model{ID: uint(random.PositiveInt(t))},
			Type:        model.EventTypeOfficial,
			Name:        reqBody.Name,
			Description: reqBody.Description,
			Location:    reqBody.Location,
			TimeStart:   reqBody.TimeStart,
			TimeEnd:     &reqBody.TimeEnd,
			CampID:      campID,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&user, nil)
		h.repo.MockEventRepository.EXPECT().
			CreateEvent(gomock.Any()).
			DoAndReturn(func(event *model.Event) error {
				event.ID = createdEvent.ID
				return nil
			})

		var eventRequest api.EventRequest

		err := eventRequest.FromOfficialEventRequest(reqBody)

		require.NoError(t, err)

		res := h.expect.POST("/api/camps/{campId}/events", campID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(eventRequest).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object()

		res.Keys().
			ContainsOnly("id", "type", "name", "description", "location", "timeStart", "timeEnd")
		res.Value("id").Number().IsEqual(createdEvent.ID)
		res.Value("type").String().IsEqual(string(model.EventTypeOfficial))
		res.Value("name").String().IsEqual(reqBody.Name)
		res.Value("description").String().IsEqual(reqBody.Description)
		res.Value("location").String().IsEqual(reqBody.Location)
		res.Value("timeStart").String().AsDateTime().IsEqual(reqBody.TimeStart)
		res.Value("timeEnd").String().AsDateTime().IsEqual(reqBody.TimeEnd)
	})

	t.Run("Failure - Non-staff user tries to create official event", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		reqBody := api.OfficialEventRequest{
			Type: api.OfficialEventRequestTypeOfficial,
			Name: random.AlphaNumericString(t, 20),
		}
		user := model.User{
			ID:      userID,
			IsStaff: false, // Non-staff user
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&user, nil)

		var eventRequest api.EventRequest

		err := eventRequest.FromOfficialEventRequest(reqBody)

		require.NoError(t, err)

		h.expect.POST("/api/camps/{campId}/events", campID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(eventRequest).
			Expect().
			Status(http.StatusForbidden)
	})
}

func TestServer_PutEvent(t *testing.T) {
	t.Parallel()

	t.Run("Success - Update event with organizer who is camp participant", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		eventID := uint(random.PositiveInt(t))
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		organizerID := random.AlphaNumericString(t, 32)
		timeStart := random.Time(t)
		timeEnd := timeStart.Add(time.Duration(random.PositiveInt(t)))
		displayColor := random.SelectFrom(
			t,
			api.DurationEventRequestDisplayColorBlue,
			api.DurationEventRequestDisplayColorGreen,
			api.DurationEventRequestDisplayColorPink,
			api.DurationEventRequestDisplayColorPurple,
			api.DurationEventRequestDisplayColorRed,
		)

		existingEvent := model.Event{
			Model:  gorm.Model{ID: eventID},
			Type:   model.EventTypeDuration,
			Name:   random.AlphaNumericString(t, 20),
			CampID: campID,
		}

		reqBody := api.DurationEventRequest{
			Type:         api.DurationEventRequestTypeDuration,
			Name:         random.AlphaNumericString(t, 20),
			Description:  random.AlphaNumericString(t, 100),
			Location:     random.AlphaNumericString(t, 50),
			TimeStart:    timeStart,
			TimeEnd:      timeEnd,
			OrganizerId:  organizerID,
			DisplayColor: displayColor,
		}

		user := model.User{
			ID:      userID,
			IsStaff: false,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&user, nil)
		h.repo.MockEventRepository.EXPECT().
			GetEventByID(eventID).
			Return(&existingEvent, nil)
		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), campID, organizerID).
			Return(true, nil)
		h.repo.MockEventRepository.EXPECT().
			UpdateEvent(gomock.Any(), eventID, gomock.Any()).
			Return(nil)

		var eventRequest api.EventRequest

		err := eventRequest.FromDurationEventRequest(reqBody)

		require.NoError(t, err)

		res := h.expect.PUT("/api/events/{eventId}", eventID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(eventRequest).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		res.Keys().
			ContainsOnly("id", "type", "name", "description", "location", "timeStart", "timeEnd", "organizerId", "displayColor")
		res.Value("id").Number().IsEqual(existingEvent.ID)
		res.Value("type").String().IsEqual(string(model.EventTypeDuration))
		res.Value("name").String().IsEqual(reqBody.Name)
		res.Value("description").String().IsEqual(reqBody.Description)
		res.Value("location").String().IsEqual(reqBody.Location)
		res.Value("timeStart").String().AsDateTime().IsEqual(reqBody.TimeStart)
		res.Value("timeEnd").String().AsDateTime().IsEqual(reqBody.TimeEnd)
		res.Value("organizerId").String().IsEqual(reqBody.OrganizerId)
		res.Value("displayColor").String().IsEqual(string(reqBody.DisplayColor))
	})

	t.Run("Failure - Update event with organizer who is not camp participant", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		eventID := uint(random.PositiveInt(t))
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		organizerID := random.AlphaNumericString(t, 32)
		timeStart := random.Time(t)
		timeEnd := timeStart.Add(time.Duration(random.PositiveInt(t)))
		existingEvent := model.Event{
			Model:  gorm.Model{ID: eventID},
			Type:   model.EventTypeDuration,
			Name:   random.AlphaNumericString(t, 20),
			CampID: campID,
		}
		reqBody := api.DurationEventRequest{
			Type:        api.DurationEventRequestTypeDuration,
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Location:    random.AlphaNumericString(t, 50),
			TimeStart:   timeStart,
			TimeEnd:     timeEnd,
			OrganizerId: organizerID,
		}
		user := model.User{
			ID:      userID,
			IsStaff: false,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&user, nil)
		h.repo.MockEventRepository.EXPECT().
			GetEventByID(eventID).
			Return(&existingEvent, nil)
		h.repo.MockCampRepository.EXPECT().
			IsCampParticipant(gomock.Any(), campID, organizerID).
			Return(false, nil)

		var eventRequest api.EventRequest

		err := eventRequest.FromDurationEventRequest(reqBody)

		require.NoError(t, err)

		h.expect.PUT("/api/events/{eventId}", eventID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(eventRequest).
			Expect().
			Status(http.StatusBadRequest).
			JSON().
			Object().
			Value("message").String().Contains("Organizer must be a participant of the camp")
	})

	t.Run("Success - Update official event by staff user", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		eventID := uint(random.PositiveInt(t))
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		timeStart := random.Time(t)
		timeEnd := timeStart.Add(time.Duration(random.PositiveInt(t)))
		existingEvent := model.Event{
			Model:  gorm.Model{ID: eventID},
			Type:   model.EventTypeOfficial,
			Name:   random.AlphaNumericString(t, 20),
			CampID: campID,
		}
		reqBody := api.OfficialEventRequest{
			Type:        api.OfficialEventRequestTypeOfficial,
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Location:    random.AlphaNumericString(t, 50),
			TimeStart:   timeStart,
			TimeEnd:     timeEnd,
		}
		user := model.User{
			ID:      userID,
			IsStaff: true, // Staff user
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&user, nil)
		h.repo.MockEventRepository.EXPECT().
			GetEventByID(eventID).
			Return(&existingEvent, nil)
		h.repo.MockEventRepository.EXPECT().
			UpdateEvent(gomock.Any(), eventID, gomock.Any()).
			Return(nil)

		var eventRequest api.EventRequest

		err := eventRequest.FromOfficialEventRequest(reqBody)

		require.NoError(t, err)

		res := h.expect.PUT("/api/events/{eventId}", eventID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(eventRequest).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		res.Keys().
			ContainsOnly("id", "type", "name", "description", "location", "timeStart", "timeEnd")
		res.Value("id").Number().IsEqual(existingEvent.ID)
		res.Value("type").String().IsEqual(string(model.EventTypeOfficial))
		res.Value("name").String().IsEqual(reqBody.Name)
		res.Value("description").String().IsEqual(reqBody.Description)
		res.Value("location").String().IsEqual(reqBody.Location)
		res.Value("timeStart").String().AsDateTime().IsEqual(reqBody.TimeStart)
		res.Value("timeEnd").String().AsDateTime().IsEqual(reqBody.TimeEnd)
	})

	t.Run("Failure - Non-staff user tries to update official event", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		eventID := uint(random.PositiveInt(t))
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		existingEvent := model.Event{
			Model:  gorm.Model{ID: eventID},
			Type:   model.EventTypeOfficial,
			Name:   random.AlphaNumericString(t, 20),
			CampID: campID,
		}
		reqBody := api.OfficialEventRequest{
			Type: api.OfficialEventRequestTypeOfficial,
			Name: random.AlphaNumericString(t, 20),
		}
		user := model.User{
			ID:      userID,
			IsStaff: false, // Non-staff user
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(&user, nil)
		h.repo.MockEventRepository.EXPECT().
			GetEventByID(eventID).
			Return(&existingEvent, nil)

		var eventRequest api.EventRequest

		err := eventRequest.FromOfficialEventRequest(reqBody)

		require.NoError(t, err)

		h.expect.PUT("/api/events/{eventId}", eventID).
			WithHeader("X-Forwarded-User", userID).
			WithJSON(eventRequest).
			Expect().
			Status(http.StatusForbidden)
	})
}

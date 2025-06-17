package router

import (
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"

	"github.com/traP-jp/rucQ/backend/model"
)

func (s *Server) GetEvents(e echo.Context, campId CampId) error {
	events, err := s.repo.GetEvents()

	if err != nil {
		e.Logger().Errorf("failed to get events: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	responseEvents := make([]EventResponse, len(events))

	for i := range events {
		var responseEvent EventResponse

		switch events[i].Type {
		case model.EventTypeDuration:
			var durationEvent DurationEventResponse

			if err := copier.Copy(&durationEvent, &events[i]); err != nil {
				e.Logger().Errorf("failed to copy duration event: %v", err)

				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}

			if err := responseEvent.FromDurationEventResponse(durationEvent); err != nil {
				e.Logger().Errorf("failed to convert duration event response: %v", err)

				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}

		case model.EventTypeMoment:
			var momentEvent MomentEventResponse

			if err := copier.Copy(&momentEvent, &events[i]); err != nil {
				e.Logger().Errorf("failed to copy moment event: %v", err)

				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}

			if err := responseEvent.FromMomentEventResponse(momentEvent); err != nil {
				e.Logger().Errorf("failed to convert moment event response: %v", err)

				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}

		case model.EventTypeOfficial:
			var officialEvent OfficialEventResponse

			if err := copier.Copy(&officialEvent, &events[i]); err != nil {
				e.Logger().Errorf("failed to copy official event: %v", err)

				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}

			if err := responseEvent.FromOfficialEventResponse(officialEvent); err != nil {
				e.Logger().Errorf("failed to convert official event response: %v", err)

				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}

		default:
			e.Logger().Errorf("unknown event type: %s", events[i].Type)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		responseEvents[i] = responseEvent
	}

	return e.JSON(http.StatusOK, responseEvents)
}

func (s *Server) PostEvent(e echo.Context, campId CampId, params PostEventParams) error {
	var req PostEventJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	var eventModel model.Event

	// Typeを取得するため、一旦DurationEventRequestを使う
	durationEventRequest, durationErr := req.AsDurationEventRequest()

	if durationErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid event request body")
	}

	switch model.EventType(durationEventRequest.Type) {
	case model.EventTypeDuration:
		if err := copier.Copy(&eventModel, &durationEventRequest); err != nil {
			e.Logger().Errorf("failed to copy duration event request to model: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

	case model.EventTypeMoment:
		momentEventRequest, err := req.AsMomentEventRequest()

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid event request body")
		}

		if err := copier.Copy(&eventModel, &momentEventRequest); err != nil {
			e.Logger().Errorf("failed to copy moment event request to model: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

	case model.EventTypeOfficial:
		officialEventRequest, err := req.AsOfficialEventRequest()

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid event request body")
		}

		user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

		if err != nil {
			e.Logger().Errorf("failed to get or create user: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		if !user.IsStaff {
			return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
		}

		if err := copier.Copy(&eventModel, &officialEventRequest); err != nil {
			e.Logger().Errorf("failed to copy official event request to model: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid event type")
	}

	eventModel.CampID = uint(campId)

	if err := s.repo.CreateEvent(&eventModel); err != nil {
		e.Logger().Errorf("failed to create event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var eventResponse EventResponse

	switch eventModel.Type {
	case model.EventTypeDuration:
		var durationEventResponse DurationEventResponse

		if err := copier.Copy(&durationEventResponse, &eventModel); err != nil {
			e.Logger().Errorf("failed to copy duration event response: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		if err := eventResponse.FromDurationEventResponse(durationEventResponse); err != nil {
			e.Logger().Errorf("failed to convert duration event response: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

	case model.EventTypeMoment:
		var momentEventResponse MomentEventResponse

		if err := copier.Copy(&momentEventResponse, &eventModel); err != nil {
			e.Logger().Errorf("failed to copy moment event response: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		if err := eventResponse.FromMomentEventResponse(momentEventResponse); err != nil {
			e.Logger().Errorf("failed to convert moment event response: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

	case model.EventTypeOfficial:
		var officialEventResponse OfficialEventResponse

		if err := copier.Copy(&officialEventResponse, &eventModel); err != nil {
			e.Logger().Errorf("failed to copy official event response: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		if err := eventResponse.FromOfficialEventResponse(officialEventResponse); err != nil {
			e.Logger().Errorf("failed to convert official event response: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

	default:
		e.Logger().Errorf("unknown event type: %s", eventModel.Type)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, &eventResponse)
}

func (s *Server) GetEvent(e echo.Context, eventID EventId) error {
	event, err := s.repo.GetEventByID(uint(eventID))
	if err != nil {
		e.Logger().Errorf("failed to get event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response EventResponse

	switch event.Type {
	case model.EventTypeDuration:
		var durationEvent DurationEventResponse

		if err := copier.Copy(&durationEvent, event); err != nil {
			e.Logger().Errorf("failed to copy duration event: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		if err := response.FromDurationEventResponse(durationEvent); err != nil {
			e.Logger().Errorf("failed to convert duration event response: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

	case model.EventTypeMoment:
		var momentEvent MomentEventResponse

		if err := copier.Copy(&momentEvent, event); err != nil {
			e.Logger().Errorf("failed to copy moment event: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		if err := response.FromMomentEventResponse(momentEvent); err != nil {
			e.Logger().Errorf("failed to convert moment event response: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

	case model.EventTypeOfficial:
		var officialEvent OfficialEventResponse

		if err := copier.Copy(&officialEvent, event); err != nil {
			e.Logger().Errorf("failed to copy official event: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		if err := response.FromOfficialEventResponse(officialEvent); err != nil {
			e.Logger().Errorf("failed to convert official event response: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

	default:
		e.Logger().Errorf("unknown event type: %s", event.Type)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &response)
}

func (s *Server) PutEvent(e echo.Context, eventID EventId, params PutEventParams) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)
	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	updateEvent, err := s.repo.GetEventByID(uint(eventID))
	if err != nil {
		e.Logger().Errorf("failed to get event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if updateEvent.Type == model.EventTypeOfficial && !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req PutEventJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	// Typeを取得するため、一旦DurationEventRequestを使う
	durationEventRequest, durationErr := req.AsDurationEventRequest()

	if durationErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid event request body")
	}

	switch model.EventType(durationEventRequest.Type) {
	case model.EventTypeDuration:
		if err := copier.Copy(updateEvent, &durationEventRequest); err != nil {
			e.Logger().Errorf("failed to copy duration event request to model: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

	case model.EventTypeMoment:
		momentEventRequest, err := req.AsMomentEventRequest()

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid event request body")
		}

		if err := copier.Copy(updateEvent, &momentEventRequest); err != nil {
			e.Logger().Errorf("failed to copy moment event request to model: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

	case model.EventTypeOfficial:
		officialEventRequest, err := req.AsOfficialEventRequest()

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid event request body")
		}

		if !user.IsStaff {
			return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
		}

		if err := copier.Copy(updateEvent, &officialEventRequest); err != nil {
			e.Logger().Errorf("failed to copy official event request to model: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid event type")
	}

	if err := s.repo.UpdateEvent(uint(eventID), updateEvent); err != nil {
		e.Logger().Errorf("failed to update event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response EventResponse

	switch updateEvent.Type {
	case model.EventTypeDuration:
		var durationEvent DurationEventResponse

		if err := copier.Copy(&durationEvent, updateEvent); err != nil {
			e.Logger().Errorf("failed to copy duration event: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		if err := response.FromDurationEventResponse(durationEvent); err != nil {
			e.Logger().Errorf("failed to convert duration event response: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

	case model.EventTypeMoment:
		var momentEvent MomentEventResponse

		if err := copier.Copy(&momentEvent, updateEvent); err != nil {
			e.Logger().Errorf("failed to copy moment event: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		if err := response.FromMomentEventResponse(momentEvent); err != nil {
			e.Logger().Errorf("failed to convert moment event response: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

	case model.EventTypeOfficial:
		var officialEvent OfficialEventResponse

		if err := copier.Copy(&officialEvent, updateEvent); err != nil {
			e.Logger().Errorf("failed to copy official event: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		if err := response.FromOfficialEventResponse(officialEvent); err != nil {
			e.Logger().Errorf("failed to convert official event response: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

	default:
		e.Logger().Errorf("unknown event type: %s", updateEvent.Type)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, response)
}

func (s *Server) DeleteEvent(e echo.Context, eventID EventId, params DeleteEventParams) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)
	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	deleteEvent, err := s.repo.GetEventByID(uint(eventID))
	if err != nil {
		e.Logger().Errorf("failed to get event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if user.ID != *deleteEvent.OrganizerID && !user.IsStaff { // イベントの主催者でない場合は削除できない
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if err := s.repo.DeleteEvent(uint(eventID)); err != nil {
		e.Logger().Errorf("failed to delete event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}

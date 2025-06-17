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
		responseEvent, err := convertModelToEventResponse(&events[i])
		if err != nil {
			e.Logger().Errorf("failed to convert event to response: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		responseEvents[i] = *responseEvent
	}

	return e.JSON(http.StatusOK, responseEvents)
}

func (s *Server) PostEvent(e echo.Context, campId CampId, params PostEventParams) error {
	var req PostEventJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	eventModel, err := convertRequestToModel(req, uint(campId))

	if err != nil {
		if httpErr, ok := err.(*echo.HTTPError); ok {
			return httpErr
		}
		e.Logger().Errorf("failed to convert request to model: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if (eventModel.Type == model.EventTypeOfficial || eventModel.Type == model.EventTypeMoment) && !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if err := s.repo.CreateEvent(eventModel); err != nil {
		e.Logger().Errorf("failed to create event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	eventResponse, err := convertModelToEventResponse(eventModel)

	if err != nil {
		if httpErr, ok := err.(*echo.HTTPError); ok {
			return httpErr
		}
		e.Logger().Errorf("failed to convert event to response: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, eventResponse)
}

func (s *Server) GetEvent(e echo.Context, eventID EventId) error {
	event, err := s.repo.GetEventByID(uint(eventID))
	if err != nil {
		e.Logger().Errorf("failed to get event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := convertModelToEventResponse(event)

	if err != nil {
		if httpErr, ok := err.(*echo.HTTPError); ok {
			return httpErr
		}
		e.Logger().Errorf("failed to convert event to response: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, response)
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

	if (updateEvent.Type == model.EventTypeOfficial || updateEvent.Type == model.EventTypeMoment) && !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req PutEventJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	if err := convertPutRequestToModel(req, updateEvent); err != nil {
		if httpErr, ok := err.(*echo.HTTPError); ok {
			return httpErr
		}
		e.Logger().Errorf("failed to convert request to model: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if (updateEvent.Type == model.EventTypeOfficial || updateEvent.Type == model.EventTypeMoment) && !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if err := s.repo.UpdateEvent(uint(eventID), updateEvent); err != nil {
		e.Logger().Errorf("failed to update event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := convertModelToEventResponse(updateEvent)

	if err != nil {
		if httpErr, ok := err.(*echo.HTTPError); ok {
			return httpErr
		}
		e.Logger().Errorf("failed to convert event to response: %v", err)
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

	if (deleteEvent.Type == model.EventTypeOfficial || deleteEvent.Type == model.EventTypeMoment) && !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if err := s.repo.DeleteEvent(uint(eventID)); err != nil {
		e.Logger().Errorf("failed to delete event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}

// convertModelToEventResponse はmodel.EventをEventResponseに変換する
func convertModelToEventResponse(event *model.Event) (*EventResponse, error) {
	var response EventResponse

	switch event.Type {
	case model.EventTypeDuration:
		var durationEvent DurationEventResponse
		if err := copier.Copy(&durationEvent, event); err != nil {
			return nil, err
		}
		if err := response.FromDurationEventResponse(durationEvent); err != nil {
			return nil, err
		}

	case model.EventTypeMoment:
		var momentEvent MomentEventResponse
		if err := copier.Copy(&momentEvent, event); err != nil {
			return nil, err
		}
		if err := response.FromMomentEventResponse(momentEvent); err != nil {
			return nil, err
		}

	case model.EventTypeOfficial:
		var officialEvent OfficialEventResponse
		if err := copier.Copy(&officialEvent, event); err != nil {
			return nil, err
		}
		if err := response.FromOfficialEventResponse(officialEvent); err != nil {
			return nil, err
		}

	default:
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "unknown event type")
	}

	return &response, nil
}

// convertRequestToModel はリクエストをmodel.Eventに変換する
func convertRequestToModel(req PostEventJSONRequestBody, campID uint) (*model.Event, error) {
	var eventModel model.Event

	// Typeを取得するため、一旦DurationEventRequestを使う
	durationEventRequest, err := req.AsDurationEventRequest()
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid event request body")
	}

	switch model.EventType(durationEventRequest.Type) {
	case model.EventTypeDuration:
		if err := copier.Copy(&eventModel, &durationEventRequest); err != nil {
			return nil, err
		}

	case model.EventTypeMoment:
		momentEventRequest, err := req.AsMomentEventRequest()
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid event request body")
		}
		if err := copier.Copy(&eventModel, &momentEventRequest); err != nil {
			return nil, err
		}

	case model.EventTypeOfficial:
		officialEventRequest, err := req.AsOfficialEventRequest()
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid event request body")
		}
		if err := copier.Copy(&eventModel, &officialEventRequest); err != nil {
			return nil, err
		}

	default:
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid event type")
	}

	eventModel.CampID = campID
	return &eventModel, nil
}

// convertPutRequestToModel はPUTリクエストでmodel.Eventを更新する
func convertPutRequestToModel(req PutEventJSONRequestBody, event *model.Event) error {
	// Typeを取得するため、一旦DurationEventRequestを使う
	durationEventRequest, durationErr := req.AsDurationEventRequest()
	if durationErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid event request body")
	}

	switch model.EventType(durationEventRequest.Type) {
	case model.EventTypeDuration:
		if err := copier.Copy(event, &durationEventRequest); err != nil {
			return err
		}

	case model.EventTypeMoment:
		momentEventRequest, err := req.AsMomentEventRequest()
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid event request body")
		}
		if err := copier.Copy(event, &momentEventRequest); err != nil {
			return err
		}

	case model.EventTypeOfficial:
		officialEventRequest, err := req.AsOfficialEventRequest()
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid event request body")
		}
		if err := copier.Copy(event, &officialEventRequest); err != nil {
			return err
		}

	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid event type")
	}

	return nil
}

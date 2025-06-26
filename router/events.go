package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"

	"github.com/traP-jp/rucQ/backend/api"
	"github.com/traP-jp/rucQ/backend/model"
)

// コンバート関数で使用するエラー変数
var (
	errUnknownEventType   = errors.New("unknown event type")
	errInvalidRequestBody = errors.New("invalid request body")
	errInvalidEventType   = errors.New("invalid event type")
)

func (s *Server) GetEvents(e echo.Context, campId api.CampId) error {
	events, err := s.repo.GetEvents()

	if err != nil {
		e.Logger().Errorf("failed to get events: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	responseEvents := make([]api.EventResponse, len(events))

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

func (s *Server) PostEvent(e echo.Context, campId api.CampId, params api.PostEventParams) error {
	var req api.PostEventJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	eventModel, err := convertEventRequestToModel(req)

	if err != nil {
		if errors.Is(err, errInvalidRequestBody) || errors.Is(err, errInvalidEventType) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		e.Logger().Errorf("failed to convert request to model: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	eventModel.CampID = uint(campId)
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
		e.Logger().Errorf("failed to convert event to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, eventResponse)
}

func (s *Server) GetEvent(e echo.Context, eventID api.EventId) error {
	event, err := s.repo.GetEventByID(uint(eventID))
	if err != nil {
		e.Logger().Errorf("failed to get event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := convertModelToEventResponse(event)

	if err != nil {
		e.Logger().Errorf("failed to convert event to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, response)
}

func (s *Server) PutEvent(e echo.Context, eventID api.EventId, params api.PutEventParams) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)
	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	existingEvent, err := s.repo.GetEventByID(uint(eventID))
	if err != nil {
		e.Logger().Errorf("failed to get event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if (existingEvent.Type == model.EventTypeOfficial || existingEvent.Type == model.EventTypeMoment) && !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.PutEventJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	newEvent, err := convertEventRequestToModel(req)

	if err != nil {
		if errors.Is(err, errInvalidRequestBody) || errors.Is(err, errInvalidEventType) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		e.Logger().Errorf("failed to convert request to model: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	newEvent.ID = existingEvent.ID

	if (newEvent.Type == model.EventTypeOfficial || newEvent.Type == model.EventTypeMoment) && !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if err := s.repo.UpdateEvent(e.Request().Context(), uint(eventID), newEvent); err != nil {
		e.Logger().Errorf("failed to update event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := convertModelToEventResponse(newEvent)

	if err != nil {
		e.Logger().Errorf("failed to convert event to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, response)
}

func (s *Server) DeleteEvent(e echo.Context, eventID api.EventId, params api.DeleteEventParams) error {
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
func convertModelToEventResponse(event *model.Event) (*api.EventResponse, error) {
	var response api.EventResponse

	switch event.Type {
	case model.EventTypeDuration:
		var durationEvent api.DurationEventResponse

		if err := copier.Copy(&durationEvent, event); err != nil {
			return nil, fmt.Errorf("failed to copy event to DurationEventResponse: %w", err)
		}

		if err := response.FromDurationEventResponse(durationEvent); err != nil {
			return nil, fmt.Errorf("failed to convert DurationEventResponse: %w", err)
		}

	case model.EventTypeMoment:
		var momentEvent api.MomentEventResponse
		if err := copier.Copy(&momentEvent, event); err != nil {
			return nil, fmt.Errorf("failed to copy event to MomentEventResponse: %w", err)
		}

		if err := response.FromMomentEventResponse(momentEvent); err != nil {
			return nil, fmt.Errorf("failed to convert MomentEventResponse: %w", err)
		}

	case model.EventTypeOfficial:
		var officialEvent api.OfficialEventResponse

		if err := copier.Copy(&officialEvent, event); err != nil {
			return nil, fmt.Errorf("failed to copy event to OfficialEventResponse: %w", err)
		}

		if err := response.FromOfficialEventResponse(officialEvent); err != nil {
			return nil, fmt.Errorf("failed to convert OfficialEventResponse: %w", err)
		}

	default:
		return nil, errUnknownEventType
	}

	return &response, nil
}

func convertEventRequestToModel(req api.EventRequest) (*model.Event, error) {
	var event model.Event

	// リクエストのTypeを取得するため、一旦DurationEventRequestを使う
	durationEventRequest, durationErr := req.AsDurationEventRequest()

	if durationErr != nil {
		return nil, errInvalidRequestBody
	}

	switch model.EventType(durationEventRequest.Type) {
	case model.EventTypeDuration:
		if err := copier.Copy(&event, &durationEventRequest); err != nil {
			return nil, fmt.Errorf("failed to copy DurationEventRequest to model.Event: %w", err)
		}

	case model.EventTypeMoment:
		momentEventRequest, err := req.AsMomentEventRequest()

		if err != nil {
			return nil, errInvalidRequestBody
		}

		if err := copier.Copy(&event, &momentEventRequest); err != nil {
			return nil, fmt.Errorf("failed to copy MomentEventRequest to model.Event: %w", err)
		}

	case model.EventTypeOfficial:
		officialEventRequest, err := req.AsOfficialEventRequest()

		if err != nil {
			return nil, errInvalidRequestBody
		}

		if err := copier.Copy(&event, &officialEventRequest); err != nil {
			return nil, fmt.Errorf("failed to copy OfficialEventRequest to model.Event: %w", err)
		}

	default:
		return nil, errInvalidEventType
	}

	return &event, nil
}

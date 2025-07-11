package router

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/model"
)

func (s *Server) GetEvents(e echo.Context, campID api.CampId) error {
	events, err := s.repo.GetEvents(e.Request().Context(), uint(campID))

	if err != nil {
		e.Logger().Errorf("failed to get events: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := converter.Convert[[]api.EventResponse](events)

	if err != nil {
		e.Logger().Errorf("failed to convert events to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, response)
}

func (s *Server) PostEvent(e echo.Context, campID api.CampId, params api.PostEventParams) error {
	var req api.PostEventJSONRequestBody

	if err := e.Bind(&req); err != nil {
		e.Logger().Warnf("failed to bind request body: %v", err)

		return err
	}

	eventModel, err := converter.Convert[model.Event](req)

	if err != nil {
		e.Logger().Errorf("failed to convert request to model: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	eventModel.CampID = uint(campID)
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if (eventModel.Type == model.EventTypeOfficial || eventModel.Type == model.EventTypeMoment) &&
		!user.IsStaff {
		e.Logger().Warnf(
			"user %s is not permitted to create event of type %s",
			*params.XForwardedUser,
			eventModel.Type,
		)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if eventModel.OrganizerID != nil {
		// TODO: ユーザーがtraQに存在するか確認する
		organizer, err := s.repo.GetOrCreateUser(e.Request().Context(), *eventModel.OrganizerID)

		if err != nil {
			e.Logger().Errorf("failed to get or create organizer user: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		// IDの正規化
		eventModel.OrganizerID = &organizer.ID
	}

	if err := s.repo.CreateEvent(&eventModel); err != nil {
		e.Logger().Errorf("failed to create event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := converter.Convert[api.EventResponse](eventModel)

	if err != nil {
		e.Logger().Errorf("failed to convert event to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, &response)
}

func (s *Server) GetEvent(e echo.Context, eventID api.EventId) error {
	event, err := s.repo.GetEventByID(uint(eventID))
	if err != nil {
		e.Logger().Errorf("failed to get event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := converter.Convert[api.EventResponse](event)

	if err != nil {
		e.Logger().Errorf("failed to convert event to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &response)
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

	if (existingEvent.Type == model.EventTypeOfficial || existingEvent.Type == model.EventTypeMoment) &&
		!user.IsStaff {
		e.Logger().Warnf(
			"user %s is not permitted to update event of type %s",
			*params.XForwardedUser,
			existingEvent.Type,
		)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.PutEventJSONRequestBody

	if err := e.Bind(&req); err != nil {
		e.Logger().Warnf("failed to bind request body: %v", err)

		return err
	}

	newEvent, err := converter.Convert[model.Event](req)

	if err != nil {
		e.Logger().Errorf("failed to convert request to model: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	newEvent.ID = existingEvent.ID

	if (newEvent.Type == model.EventTypeOfficial || newEvent.Type == model.EventTypeMoment) &&
		!user.IsStaff {
		e.Logger().Warnf(
			"user %s is not permitted to update event to type %s",
			*params.XForwardedUser,
			newEvent.Type,
		)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if err := s.repo.UpdateEvent(e.Request().Context(), uint(eventID), &newEvent); err != nil {
		e.Logger().Errorf("failed to update event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := converter.Convert[api.EventResponse](newEvent)

	if err != nil {
		e.Logger().Errorf("failed to convert event to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &response)
}

func (s *Server) DeleteEvent(
	e echo.Context,
	eventID api.EventId,
	params api.DeleteEventParams,
) error {
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

	if (deleteEvent.Type == model.EventTypeOfficial || deleteEvent.Type == model.EventTypeMoment) &&
		!user.IsStaff {
		e.Logger().Warnf(
			"user %s is not permitted to delete event of type %s",
			*params.XForwardedUser,
			deleteEvent.Type,
		)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if err := s.repo.DeleteEvent(uint(eventID)); err != nil {
		e.Logger().Errorf("failed to delete event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}

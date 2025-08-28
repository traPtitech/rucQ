package router

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/model"
)

func (s *Server) GetEvents(e echo.Context, campID api.CampId) error {
	events, err := s.repo.GetEvents(e.Request().Context(), uint(campID))

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get events",
			slog.String("error", err.Error()),
			slog.Int("campId", int(campID)),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := converter.Convert[[]api.EventResponse](events)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert events to response",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, response)
}

func (s *Server) PostEvent(e echo.Context, campID api.CampId, params api.PostEventParams) error {
	var req api.PostEventJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	eventModel, err := converter.Convert[model.Event](req)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request to model",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	eventModel.CampID = uint(campID)
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if (eventModel.Type == model.EventTypeOfficial || eventModel.Type == model.EventTypeMoment) &&
		!user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if eventModel.OrganizerID != nil {
		isCampParticipant, err := s.repo.IsCampParticipant(
			e.Request().Context(),
			uint(campID),
			*eventModel.OrganizerID,
		)

		if err != nil {
			slog.ErrorContext(
				e.Request().Context(),
				"failed to get or create organizer user",
				slog.String("error", err.Error()),
			)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		if !isCampParticipant {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				"Organizer must be a participant of the camp",
			)
		}
	}

	if err := s.repo.CreateEvent(&eventModel); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to create event",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := converter.Convert[api.EventResponse](eventModel)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert event to response",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, &response)
}

func (s *Server) GetEvent(e echo.Context, eventID api.EventId) error {
	event, err := s.repo.GetEventByID(uint(eventID))
	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get event",
			slog.String("error", err.Error()),
			slog.Int("eventId", int(eventID)),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := converter.Convert[api.EventResponse](event)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert event to response",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &response)
}

func (s *Server) PutEvent(e echo.Context, eventID api.EventId, params api.PutEventParams) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)
	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	existingEvent, err := s.repo.GetEventByID(uint(eventID))
	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get event",
			slog.String("error", err.Error()),
			slog.Int("eventId", int(eventID)),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if (existingEvent.Type == model.EventTypeOfficial || existingEvent.Type == model.EventTypeMoment) &&
		!user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.PutEventJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	newEvent, err := converter.Convert[model.Event](req)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request to model",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	newEvent.ID = existingEvent.ID

	if (newEvent.Type == model.EventTypeOfficial || newEvent.Type == model.EventTypeMoment) &&
		!user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if newEvent.OrganizerID != nil {
		isCampParticipant, err := s.repo.IsCampParticipant(
			e.Request().Context(),
			existingEvent.CampID,
			*newEvent.OrganizerID,
		)

		if err != nil {
			slog.ErrorContext(
				e.Request().Context(),
				"failed to check if organizer is a camp participant",
				slog.String("error", err.Error()),
			)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		if !isCampParticipant {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				"Organizer must be a participant of the camp",
			)
		}
	}

	if err := s.repo.UpdateEvent(e.Request().Context(), uint(eventID), &newEvent); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to update event",
			slog.String("error", err.Error()),
			slog.Int("eventId", int(eventID)),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := converter.Convert[api.EventResponse](newEvent)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert event to response",
			slog.String("error", err.Error()),
		)

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
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	deleteEvent, err := s.repo.GetEventByID(uint(eventID))
	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get event",
			slog.String("error", err.Error()),
			slog.Int("eventId", int(eventID)),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if (deleteEvent.Type == model.EventTypeOfficial || deleteEvent.Type == model.EventTypeMoment) &&
		!user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if err := s.repo.DeleteEvent(uint(eventID)); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to delete event",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}

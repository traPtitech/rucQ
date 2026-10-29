package router

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
)

func (s *Server) GetEvents(e echo.Context, campID api.CampId) error {
	events, err := s.repo.GetEvents(e.Request().Context(), uint(campID))

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get events: %w", err))
	}

	response, err := converter.Convert[[]api.EventResponse](events)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert events to response: %w", err))
	}

	return e.JSON(http.StatusOK, response)
}

func (s *Server) PostEvent(e echo.Context, campID api.CampId, params api.PostEventParams) error {
	camp, err := s.repo.GetCampByID(e.Request().Context(), uint(campID))
	if err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get camp: %w", err))
	}

	if time.Now().After(camp.DateEnd) {
		return echo.NewHTTPError(http.StatusBadRequest, "Camp has already ended")
	}

	var req api.PostEventJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	eventModel, err := converter.Convert[model.Event](req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert request to model: %w", err))
	}

	eventModel.CampID = uint(campID)
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
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
			return echo.NewHTTPError(http.StatusInternalServerError).
				SetInternal(fmt.Errorf("failed to get or create organizer user: %w", err))
		}

		if !isCampParticipant {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				"Organizer must be a participant of the camp",
			)
		}
	}

	if err := s.repo.CreateEvent(&eventModel); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to create event: %w", err))
	}

	response, err := converter.Convert[api.EventResponse](eventModel)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert event to response: %w", err))
	}

	return e.JSON(http.StatusCreated, &response)
}

func (s *Server) GetEvent(e echo.Context, eventID api.EventId) error {
	event, err := s.repo.GetEventByID(uint(eventID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get event (eventId: %d): %w", eventID, err))
	}

	response, err := converter.Convert[api.EventResponse](event)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert event to response: %w", err))
	}

	return e.JSON(http.StatusOK, &response)
}

func (s *Server) PutEvent(e echo.Context, eventID api.EventId, params api.PutEventParams) error {
	existingEvent, err := s.repo.GetEventByID(uint(eventID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get event (eventId: %d): %w", eventID, err))
	}

	camp, err := s.repo.GetCampByID(e.Request().Context(), existingEvent.CampID)
	if err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get camp: %w", err))
	}

	if time.Now().After(camp.DateEnd) {
		return echo.NewHTTPError(http.StatusBadRequest, "Camp has already ended")
	}

	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user (userId: %s): %w", *params.XForwardedUser, err))
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert request to model: %w", err))
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
			return echo.NewHTTPError(http.StatusInternalServerError).
				SetInternal(fmt.Errorf("failed to check if organizer is a camp participant: %w", err))
		}

		if !isCampParticipant {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				"Organizer must be a participant of the camp",
			)
		}
	}

	if err := s.repo.UpdateEvent(e.Request().Context(), uint(eventID), &newEvent); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to update event (eventId: %d): %w", eventID, err))
	}

	response, err := converter.Convert[api.EventResponse](newEvent)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert event to response: %w", err))
	}

	return e.JSON(http.StatusOK, &response)
}

func (s *Server) DeleteEvent(
	e echo.Context,
	eventID api.EventId,
	params api.DeleteEventParams,
) error {
	deleteEvent, err := s.repo.GetEventByID(uint(eventID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get event (eventId: %d): %w", eventID, err))
	}

	camp, err := s.repo.GetCampByID(e.Request().Context(), deleteEvent.CampID)
	if err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get camp: %w", err))
	}

	if time.Now().After(camp.DateEnd) {
		return echo.NewHTTPError(http.StatusBadRequest, "Camp has already ended")
	}

	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user (userId: %s): %w", *params.XForwardedUser, err))
	}

	if (deleteEvent.Type == model.EventTypeOfficial || deleteEvent.Type == model.EventTypeMoment) &&
		!user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if err := s.repo.DeleteEvent(uint(eventID)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to delete event: %w", err))
	}

	return e.NoContent(http.StatusNoContent)
}

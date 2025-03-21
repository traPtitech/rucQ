package handler

import (
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
	"github.com/traP-jp/rucQ/backend/model"
)

const trapTraqID = "traP"

func (s *Server) GetEvents(e echo.Context) error {
	events, err := s.repo.GetEvents()

	if err != nil {
		e.Logger().Errorf("failed to get events: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response []Event

	if err := copier.Copy(&response, &events); err != nil {
		e.Logger().Errorf("failed to copy events: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, response)
}

func (s *Server) PostEvent(e echo.Context, params PostEventParams) error {
	var req PostEventJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	var eventModel model.Event

	if err := copier.Copy(&eventModel, &req); err != nil {
		e.Logger().Errorf("failed to copy request to model: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	organizerTraqID := params.XForwardedUser

	if req.CreateAsStaff {
		user, err := s.repo.GetOrCreateUser(*organizerTraqID)

		if err != nil {
			e.Logger().Errorf("failed to get or create user: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		if !user.IsStaff {
			return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
		}

		trapTraqID := "traP"
		organizerTraqID = &trapTraqID
		eventModel.ByStaff = true
	}

	eventModel.OrganizerTraqID = *organizerTraqID

	if err := s.repo.CreateEvent(&eventModel); err != nil {
		e.Logger().Errorf("failed to create event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var eventResponse Event

	if err := copier.Copy(&eventResponse, &eventModel); err != nil {
		e.Logger().Errorf("failed to copy model to response: %v", err)

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

	var response Event

	if err := copier.Copy(&response, event); err != nil {
		e.Logger().Errorf("failed to copy event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &response)
}

func (s *Server) PutEvent(e echo.Context, eventID EventId, params PutEventParams) error {
	user, err := s.repo.GetOrCreateUser(*params.XForwardedUser)
	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	updateEvent, err := s.repo.GetEventByID(uint(eventID))
	if err != nil {
		e.Logger().Errorf("failed to get event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if user.TraqID != updateEvent.OrganizerTraqID && !user.IsStaff { // イベントの主催者orスタッフでない場合は更新できない
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req PutEventJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	if err := copier.Copy(updateEvent, &req); err != nil {
		e.Logger().Errorf("failed to copy request to model: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.UpdateEvent(uint(eventID), updateEvent); err != nil {
		e.Logger().Errorf("failed to update event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response Event

	if err := copier.Copy(&response, updateEvent); err != nil {
		e.Logger().Errorf("failed to copy model to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &response)
}

func (s *Server) DeleteEvent(e echo.Context, eventID EventId, params DeleteEventParams) error {
	user, err := s.repo.GetOrCreateUser(*params.XForwardedUser)
	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	deleteEvent, err := s.repo.GetEventByID(uint(eventID))
	if err != nil {
		e.Logger().Errorf("failed to get event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if user.TraqID != deleteEvent.OrganizerTraqID && !user.IsStaff { // イベントの主催者でない場合は削除できない
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if err := s.repo.DeleteEvent(uint(eventID)); err != nil {
		e.Logger().Errorf("failed to delete event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}

package handler

import (
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
)

func (s *Server) GetParticipants(e echo.Context, eventID EventId) error {
	participants, err := s.repo.GetParticipants(uint(eventID))

	if err != nil {
		e.Logger().Errorf("failed to get participants: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response []User

	if err := copier.Copy(&response, participants); err != nil {
		e.Logger().Errorf("failed to copy participants: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, response)
}

func (s *Server) UnregisterEvent(e echo.Context, eventID EventId, params UnregisterEventParams) error {
	user, err := s.repo.GetOrCreateUser(*params.XForwardedUser)
	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.UnregisterEvent(uint(eventID), user); err != nil {
		e.Logger().Errorf("failed to unregister event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}

func (s *Server) RegisterEvent(e echo.Context, eventID EventId, params RegisterEventParams) error {
	user, err := s.repo.GetOrCreateUser(*params.XForwardedUser)
	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.RegisterEvent(uint(eventID), user); err != nil {
		e.Logger().Errorf("failed to register event: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response User

	if err := copier.Copy(&response, user); err != nil {
		e.Logger().Errorf("failed to copy user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, &response)
}

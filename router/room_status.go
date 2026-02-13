package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
)

func (s *Server) PutRoomStatus(
	e echo.Context,
	roomID api.RoomId,
	params api.PutRoomStatusParams,
) error {
	var req api.PutRoomStatusJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	campID, err := s.repo.GetRoomCampID(e.Request().Context(), uint(roomID))
	if err != nil {
		if errors.Is(err, repository.ErrRoomNotFound) {
			return echo.ErrNotFound
		}
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get room camp id: %w", err))
	}

	isParticipant, err := s.repo.IsCampParticipant(
		e.Request().Context(),
		campID,
		*params.XForwardedUser,
	)
	if err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			return echo.ErrNotFound
		}
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to check camp participation: %w", err))
	}

	if !isParticipant {
		return echo.ErrForbidden
	}

	_, err = s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
	}

	status, err := converter.Convert[model.RoomStatus](req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert request to room status: %w", err))
	}

	if err := s.repo.SetRoomStatus(
		e.Request().Context(),
		uint(roomID),
		&status,
		*params.XForwardedUser,
	); err != nil {
		if errors.Is(err, repository.ErrRoomNotFound) {
			return echo.ErrNotFound
		}
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to set room status: %w", err))
	}

	return e.NoContent(http.StatusNoContent)
}

func (s *Server) GetRoomStatusLogs(
	e echo.Context,
	roomID api.RoomId,
) error {
	logs, err := s.repo.GetRoomStatusLogs(e.Request().Context(), uint(roomID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get room status logs: %w", err))
	}

	res, err := converter.Convert[[]api.RoomStatusLog](logs)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert room status logs: %w", err))
	}

	return e.JSON(http.StatusOK, res)
}

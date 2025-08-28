package router

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
)

func (s *Server) AdminPostRoom(e echo.Context, params api.AdminPostRoomParams) error {
	operator, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !operator.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPostRoomJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	roomModel, err := converter.Convert[model.Room](req)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request to model",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.CreateRoom(e.Request().Context(), &roomModel); err != nil {
		if errors.Is(err, repository.ErrUserOrRoomGroupNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid user or room group ID")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to create room",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// MemberのisStaffなどを正しく返すために取得
	updatedRoom, err := s.repo.GetRoomByID(roomModel.ID)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get room by ID",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.RoomResponse](updatedRoom)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert model to response",
			slog.String("error", err.Error()),
			slog.Int("roomId", int(updatedRoom.ID)),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, res)
}

func (s *Server) AdminPutRoom(
	e echo.Context,
	roomID api.RoomId,
	params api.AdminPutRoomParams,
) error {
	operator, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !operator.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPutRoomJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	roomModel, err := converter.Convert[model.Room](req)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request to model",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.UpdateRoom(e.Request().Context(), uint(roomID), &roomModel); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "User not found")
		}

		if errors.Is(err, repository.ErrRoomGroupNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "Room group not found")
		}

		if errors.Is(err, repository.ErrRoomNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Room not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to update room",
			slog.String("error", err.Error()),
			slog.Int("roomId", roomID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// MemberのisStaffなどを正しく返すために取得
	updatedRoom, err := s.repo.GetRoomByID(uint(roomID))

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get room by ID",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.RoomResponse](updatedRoom)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert model to response",
			slog.String("error", err.Error()),
			slog.Int("roomId", int(updatedRoom.ID)),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}

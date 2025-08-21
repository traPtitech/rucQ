package router

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jinzhu/copier"
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
		slog.WarnContext(
			e.Request().Context(),
			"user is not a staff member",
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPostRoomJSONRequestBody

	if err := e.Bind(&req); err != nil {
		slog.WarnContext(
			e.Request().Context(),
			"failed to bind request body",
			slog.String("error", err.Error()),
		)

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
			slog.WarnContext(
				e.Request().Context(),
				"user or room group not found",
				slog.Int("roomGroupId", int(roomModel.RoomGroupID)),
			)

			return echo.NewHTTPError(http.StatusBadRequest, "Invalid user or room group ID")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to create room",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var res api.RoomResponse

	if err := copier.Copy(&res, &roomModel); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to copy model to response",
			slog.String("error", err.Error()),
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
		slog.WarnContext(
			e.Request().Context(),
			"user is not a staff member",
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPutRoomJSONRequestBody

	if err := e.Bind(&req); err != nil {
		slog.WarnContext(
			e.Request().Context(),
			"failed to bind request body",
			slog.String("error", err.Error()),
		)

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
		slog.ErrorContext(
			e.Request().Context(),
			"failed to update room",
			slog.String("error", err.Error()),
			slog.Int("roomId", roomID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var res api.RoomResponse

	if err := copier.Copy(&res, roomModel); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to copy model to response",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}

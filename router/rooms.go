package router

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
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

	var roomModel model.Room

	if err := copier.Copy(&roomModel, &req); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to copy request to model",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	roomModel.Members = make([]model.User, len(req.MemberIds))

	for i := range req.MemberIds {
		member, err := s.repo.GetOrCreateUser(e.Request().Context(), req.MemberIds[i])

		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "User not found")
			}

			slog.ErrorContext(
				e.Request().Context(),
				"failed to get user",
				slog.String("error", err.Error()),
				slog.String("userId", req.MemberIds[i]),
			)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		roomModel.Members[i] = *member
	}

	if err := s.repo.CreateRoom(&roomModel); err != nil {
		if errors.Is(err, model.ErrAlreadyExists) {
			return echo.NewHTTPError(http.StatusConflict, "Room already exists")
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
	roomId api.RoomId,
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

	roomModel, err := s.repo.GetRoomByID(uint(roomId))

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"room not found",
				slog.Int("roomId", int(roomId)),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Room not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to get room",
			slog.String("error", err.Error()),
			slog.Int("roomId", int(roomId)),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := copier.Copy(roomModel, &req); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to copy request to model",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	roomModel.Members = make([]model.User, len(req.MemberIds))

	for i := range req.MemberIds {
		member, err := s.repo.GetOrCreateUser(e.Request().Context(), req.MemberIds[i])

		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "User not found")
			}

			slog.ErrorContext(
				e.Request().Context(),
				"failed to get user",
				slog.String("error", err.Error()),
				slog.String("userId", req.MemberIds[i]),
			)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		roomModel.Members[i] = *member
	}

	if err := s.repo.UpdateRoom(roomModel); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to update room",
			slog.String("error", err.Error()),
			slog.Int("roomId", int(roomId)),
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

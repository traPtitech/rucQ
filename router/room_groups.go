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

func (s *Server) AdminPostRoomGroup(
	e echo.Context,
	campID api.CampId,
	params api.AdminPostRoomGroupParams,
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

	if !user.IsStaff {
		slog.WarnContext(
			e.Request().Context(),
			"non-staff user attempted to create room group",
			slog.String("userId", *params.XForwardedUser),
			slog.Int("campID", campID),
		)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPostRoomGroupJSONRequestBody

	if err := e.Bind(&req); err != nil {
		slog.WarnContext(
			e.Request().Context(),
			"failed to bind request body",
			slog.String("error", err.Error()),
		)

		return err
	}

	roomGroup, err := converter.Convert[model.RoomGroup](req)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request body",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	roomGroup.CampID = uint(campID)

	if err := s.repo.CreateRoomGroup(e.Request().Context(), &roomGroup); err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"camp not found",
				slog.Int("campID", campID),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to create room group",
			slog.String("error", err.Error()),
			slog.Int("campID", campID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.RoomGroupResponse](roomGroup)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert response body",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, res)
}

func (s *Server) AdminPutRoomGroup(
	e echo.Context,
	roomGroupID api.RoomGroupId,
	params api.AdminPutRoomGroupParams,
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

	if !user.IsStaff {
		slog.WarnContext(
			e.Request().Context(),
			"non-staff user attempted to update room group",
			slog.String("userId", *params.XForwardedUser),
			slog.Int("roomGroupID", roomGroupID),
		)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPutRoomGroupJSONRequestBody

	if err := e.Bind(&req); err != nil {
		slog.WarnContext(
			e.Request().Context(),
			"failed to bind request body",
			slog.String("error", err.Error()),
		)

		return err
	}

	roomGroup, err := converter.Convert[model.RoomGroup](req)
	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request body",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.UpdateRoomGroup(e.Request().Context(), uint(roomGroupID), &roomGroup); err != nil {
		if errors.Is(err, repository.ErrRoomGroupNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"room group not found",
				slog.Int("roomGroupID", roomGroupID),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Room group not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to update room group",
			slog.String("error", err.Error()),
			slog.Int("roomGroupID", roomGroupID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.RoomGroupResponse](roomGroup)
	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert response body",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}

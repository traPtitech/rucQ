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

func (s *Server) GetRollCalls(e echo.Context, campID api.CampId) error {
	rollCalls, err := s.repo.GetRollCalls(e.Request().Context(), uint(campID))

	if err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"camp not found when getting roll calls",
				slog.Int("campId", campID),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to get roll calls",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[[]api.RollCallResponse](rollCalls)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert roll calls",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}

func (s *Server) AdminPostRollCall(
	e echo.Context,
	campID api.CampId,
	params api.AdminPostRollCallParams,
) error {
	if params.XForwardedUser == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "X-Forwarded-User header is required")
	}

	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPostRollCallJSONRequestBody

	if err := e.Bind(&req); err != nil {
		slog.WarnContext(
			e.Request().Context(),
			"failed to bind request body",
			slog.String("error", err.Error()),
		)

		return err
	}

	rollCall, err := converter.Convert[model.RollCall](req)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request body",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	rollCall.CampID = uint(campID)

	if err := s.repo.CreateRollCall(e.Request().Context(), &rollCall); err != nil {
		if errors.Is(err, repository.ErrCampNotFound) ||
			errors.Is(err, repository.ErrUserNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"failed to create roll call",
				slog.String("error", err.Error()),
				slog.Int("campId", campID),
			)

			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to create roll call",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.RollCallResponse](rollCall)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert roll call",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, res)
}

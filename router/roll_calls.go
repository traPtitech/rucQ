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

func (s *Server) GetRollCalls(e echo.Context, campID api.CampId) error {
	rollCalls, err := s.repo.GetRollCalls(e.Request().Context(), uint(campID))

	if err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get roll calls: %w", err))
	}

	res, err := converter.Convert[[]api.RollCallResponse](rollCalls)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert roll calls: %w", err))
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPostRollCallJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	rollCall, err := converter.Convert[model.RollCall](req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert request body: %w", err))
	}

	rollCall.CampID = uint(campID)

	if err := s.repo.CreateRollCall(e.Request().Context(), &rollCall); err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		if errors.Is(err, repository.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "One or more subject users not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to create roll call: %w", err))
	}

	res, err := converter.Convert[api.RollCallResponse](rollCall)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert roll call: %w", err))
	}

	return e.JSON(http.StatusCreated, res)
}

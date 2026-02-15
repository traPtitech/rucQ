package router

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
)

func (s *Server) GetActivities(
	e echo.Context,
	campID api.CampId,
	params api.GetActivitiesParams,
) error {
	if params.XForwardedUser == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "X-Forwarded-User header is required")
	}

	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
	}

	activities, err := s.activityService.GetActivities(
		e.Request().Context(),
		uint(campID),
		user.ID,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get activities: %w", err))
	}

	response, err := converter.ConvertActivityResponses(activities)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert activities to response: %w", err))
	}

	return e.JSON(http.StatusOK, response)
}

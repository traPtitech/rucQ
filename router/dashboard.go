package router

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
)

func (s *Server) GetDashboard(
	e echo.Context,
	campID api.CampId,
	params api.GetDashboardParams,
) error {
	isParticipant, err := s.repo.IsCampParticipant(
		e.Request().Context(),
		uint(campID),
		*params.XForwardedUser,
	)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			e.Logger().Warnf("camp with ID %d not found", campID)

			return echo.NewHTTPError(
				http.StatusNotFound,
				"Camp not found",
			)
		}

		e.Logger().Errorf("Failed to check camp participation: %v", err)

		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Internal server error",
		)
	}

	if !isParticipant {
		e.Logger().Warnf("user %s is not a participant of camp %d", *params.XForwardedUser, campID)

		return echo.NewHTTPError(http.StatusNotFound, "User is not a participant of this camp")
	}

	// TODO: ユーザーのPaymentとRoomを取得してレスポンスに含める
	res := api.DashboardResponse{
		Id: *params.XForwardedUser,
	}

	return e.JSON(http.StatusOK, &res)
}

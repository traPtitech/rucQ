package router

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
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
		e.Logger().Errorf("Failed to check camp participation: %v", err)

		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Internal server error",
		)
	}

	if !isParticipant {
		return echo.NewHTTPError(http.StatusNotFound, "User is not a participant of this camp")
	}

	// TODO: ユーザーのPaymentとRoomを取得してレスポンスに含める
	res := api.DashboardResponse{
		Id: *params.XForwardedUser,
	}

	return e.JSON(http.StatusOK, &res)
}

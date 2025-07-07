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
	// ユーザーがキャンプの参加者かどうかを確認
	isParticipant, err := s.repo.IsCampParticipant(
		e.Request().Context(),
		uint(campID),
		*params.XForwardedUser,
	)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Failed to check camp participation",
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

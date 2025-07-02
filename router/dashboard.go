package router

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
)

func (s *Server) GetDashboard(e echo.Context, _ api.CampId, params api.GetDashboardParams) error {
	// TODO: ユーザーのPaymentとRoomを取得してレスポンスに含める
	res := api.DashboardResponse{
		Id: *params.XForwardedUser,
	}

	return e.JSON(http.StatusOK, &res)
}

package router

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traP-jp/rucQ/backend/api"
)

func (s *Server) AdminPutOption(e echo.Context, optionId api.OptionId, params api.AdminPutOptionParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "Not implemented")
}

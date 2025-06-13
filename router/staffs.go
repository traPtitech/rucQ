package router

import (
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
)

func (s *Server) GetStaffs(e echo.Context) error {
	staffs, err := s.repo.GetStaffs()

	if err != nil {
		e.Logger().Errorf("failed to get staffs: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response []UserResponse

	if err := copier.Copy(&response, &staffs); err != nil {
		e.Logger().Errorf("failed to copy staffs: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &response)
}

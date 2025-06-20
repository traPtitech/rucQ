package router

import (
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
	"github.com/traP-jp/rucQ/backend/model"
)

func (s *Server) AdminPostOption(e echo.Context, params AdminPostOptionParams) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req AdminPostOptionJSONRequestBody

	if err := e.Bind(&req); err != nil {
		e.Logger().Errorf("failed to bind request body: %v", err)

		return echo.NewHTTPError(http.StatusBadRequest, "Bad request")
	}

	var option model.Option

	if err := copier.Copy(&option, &req); err != nil {
		e.Logger().Errorf("failed to copy request body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.CreateOption(&option); err != nil {
		e.Logger().Errorf("failed to create option: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var res OptionResponse

	if err := copier.Copy(&res, &option); err != nil {
		e.Logger().Errorf("failed to copy response body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, res)
}

func (s *Server) AdminPutOption(e echo.Context, optionId OptionId, params AdminPutOptionParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "Not implemented")
}

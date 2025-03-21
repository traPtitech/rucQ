package handler

import (
	"net/http"
	"os"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
)

func (s *Server) GetStaffs(e echo.Context) error {
	staffs, err := s.repo.GetStaffs()

	if err != nil {
		e.Logger().Errorf("failed to get staffs: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response []User

	if err := copier.Copy(&response, &staffs); err != nil {
		e.Logger().Errorf("failed to copy staffs: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &response)
}

func (s *Server) PostStaff(e echo.Context, params PostStaffParams) error {
	if os.Getenv("RUCQ_DEBUG") != "true" {
		loggedInUser, err := s.repo.GetOrCreateUser(*params.XForwardedUser)

		if err != nil {
			e.Logger().Errorf("failed to get or create user: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		if !loggedInUser.IsStaff {
			return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
		}
	}

	var req PostStaffJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	newStaff, err := s.repo.GetOrCreateUser(req.TraqId)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.SetUserIsStaff(newStaff, true); err != nil {
		e.Logger().Errorf("failed to set user as staff: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}

func (s *Server) DeleteStaff(e echo.Context, params DeleteStaffParams) error {
	loggedInUser, err := s.repo.GetOrCreateUser(*params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !loggedInUser.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	staff, err := s.repo.GetOrCreateUser(params.StaffId)

	if err != nil {
		e.Logger().Errorf("failed to get or create staff: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.SetUserIsStaff(staff, false); err != nil {
		e.Logger().Errorf("failed to set user as staff: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}

package router

import (
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
)

func (s *Server) GetMe(e echo.Context, params api.GetMeParams) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response api.UserResponse

	if err := copier.Copy(&response, &user); err != nil {
		e.Logger().Errorf("failed to copy user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &response)
}

func (s *Server) GetStaffs(e echo.Context) error {
	staffs, err := s.repo.GetStaffs()

	if err != nil {
		e.Logger().Errorf("failed to get staffs: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response []api.UserResponse

	if err := copier.Copy(&response, &staffs); err != nil {
		e.Logger().Errorf("failed to copy staffs: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &response)
}

// AdminPutUser ユーザー情報を更新（管理者用）
func (s *Server) AdminPutUser(
	e echo.Context,
	targetUserID string,
	params api.AdminPutUserParams,
) error {
	operator, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create operator user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// 開発環境では管理者でなくてもユーザー情報を更新できるようにする
	if !s.isDev && !operator.IsStaff {
		e.Logger().Warnf("user %s is not a staff member", *params.XForwardedUser)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	targetUser, err := s.repo.GetOrCreateUser(e.Request().Context(), targetUserID)

	if err != nil {
		e.Logger().Errorf("failed to get or create target user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var req api.AdminPutUserJSONRequestBody

	if err := e.Bind(&req); err != nil {
		e.Logger().Warnf("failed to bind request: %v", err)

		return err
	}

	if err := copier.Copy(targetUser, &req); err != nil {
		e.Logger().Errorf("failed to copy request to target user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// TODO: Not foundの場合のエラーハンドリングを追加
	if err := s.repo.UpdateUser(e.Request().Context(), targetUser); err != nil {
		e.Logger().Errorf("failed to update user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response api.UserResponse

	if err := copier.Copy(&response, targetUser); err != nil {
		e.Logger().Errorf("failed to copy updated user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &response)
}

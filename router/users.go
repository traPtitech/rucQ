package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/service/traq"
)

func (s *Server) GetMe(e echo.Context, params api.GetMeParams) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
	}

	var response api.UserResponse

	if err := copier.Copy(&response, &user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to copy user: %w", err))
	}

	return e.JSON(http.StatusOK, &response)
}

func (s *Server) AdminGetUser(
	e echo.Context,
	userID api.UserId,
	params api.AdminGetUserParams,
) error {
	operator, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		return echo.ErrInternalServerError.
			SetInternal(fmt.Errorf("failed to get or create operator: %w", err))
	}

	if !operator.IsStaff {
		return echo.ErrForbidden
	}

	targetUserID, err := s.traqService.GetCanonicalUserName(e.Request().Context(), userID)

	if err != nil {
		if errors.Is(err, traq.ErrUserNotFound) {
			return echo.ErrNotFound
		}

		return echo.ErrInternalServerError.SetInternal(
			fmt.Errorf("failed to get canonical user name: %w", err),
		)
	}

	targetUser, err := s.repo.GetOrCreateUser(e.Request().Context(), targetUserID)

	if err != nil {
		return echo.ErrInternalServerError.SetInternal(
			fmt.Errorf("failed to get or create target user: %w", err),
		)
	}

	res, err := converter.Convert[api.UserResponse](targetUser)

	if err != nil {
		return echo.ErrInternalServerError.SetInternal(
			fmt.Errorf("failed to convert target user: %w", err),
		)
	}

	return e.JSON(http.StatusOK, &res)
}

func (s *Server) GetStaffs(e echo.Context) error {
	staffs, err := s.repo.GetStaffs()

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get staffs: %w", err))
	}

	var response []api.UserResponse

	if err := copier.Copy(&response, &staffs); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to copy staffs: %w", err))
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
	}

	// 開発環境では管理者でなくてもユーザー情報を更新できるようにする
	if !s.isDev && !operator.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	targetUser, err := s.repo.GetOrCreateUser(e.Request().Context(), targetUserID)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create target user: %w", err))
	}

	var req api.AdminPutUserJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	if err := copier.Copy(targetUser, &req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to copy request to target user: %w", err))
	}

	// TODO: Not foundの場合のエラーハンドリングを追加
	if err := s.repo.UpdateUser(e.Request().Context(), targetUser); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to update user: %w", err))
	}

	var response api.UserResponse

	if err := copier.Copy(&response, targetUser); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to copy updated user: %w", err))
	}

	return e.JSON(http.StatusOK, &response)
}

package router

import (
	"log/slog"
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
)

func (s *Server) GetMe(e echo.Context, params api.GetMeParams) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response api.UserResponse

	if err := copier.Copy(&response, &user); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to copy user",
			slog.String("error", err.Error()),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &response)
}

func (s *Server) GetStaffs(e echo.Context) error {
	staffs, err := s.repo.GetStaffs()

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get staffs",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response []api.UserResponse

	if err := copier.Copy(&response, &staffs); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to copy staffs",
			slog.String("error", err.Error()),
		)

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
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// 開発環境では管理者でなくてもユーザー情報を更新できるようにする
	if !s.isDev && !operator.IsStaff {
		slog.WarnContext(
			e.Request().Context(),
			"user is not a staff member",
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	targetUser, err := s.repo.GetOrCreateUser(e.Request().Context(), targetUserID)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
			slog.String("userId", targetUserID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var req api.AdminPutUserJSONRequestBody

	if err := e.Bind(&req); err != nil {
		slog.WarnContext(
			e.Request().Context(),
			"failed to bind request",
			slog.String("error", err.Error()),
		)

		return err
	}

	if err := copier.Copy(targetUser, &req); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to copy request to target user",
			slog.String("error", err.Error()),
			slog.String("userId", targetUserID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// TODO: Not foundの場合のエラーハンドリングを追加
	if err := s.repo.UpdateUser(e.Request().Context(), targetUser); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to update user",
			slog.String("error", err.Error()),
			slog.String("userId", targetUserID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response api.UserResponse

	if err := copier.Copy(&response, targetUser); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to copy updated user",
			slog.String("error", err.Error()),
			slog.String("userId", targetUserID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &response)
}

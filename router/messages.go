package router

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
)

// AdminPostMessage は DM を送信するハンドラです。
func (s *Server) AdminPostMessage(
	e echo.Context,
	userID api.UserId,
	params api.AdminPostMessageParams,
) error {
	var req api.AdminPostMessageJSONRequestBody
	if err := e.Bind(&req); err != nil {
		e.Logger().Warnf("failed to bind request: %v", err)
		return err
	}

	// スタッフだけがbotを用いてdmを送信できるようにする
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)
	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// スタッフじゃなければはじく
	if !user.IsStaff {
		e.Logger().Warnf("user %s is not a staff member", *params.XForwardedUser)
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	// 指定時刻まで待機してからDMを送信する
	go func() {
		if !req.SendAt.IsZero() {
			time.Sleep(time.Until(req.SendAt))
		}

		err := s.traqService.PostDirectMessage(context.Background(), string(userID), req.Content)
		if err != nil {
			e.Logger().Errorf("failed to send direct message to %s: %v", userID, err)
		}
	}()

	return e.NoContent(http.StatusAccepted)
}

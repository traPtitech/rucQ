package router

import (
	"context"
	"log/slog"
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
		slog.WarnContext(
			e.Request().Context(),
			"failed to bind request",
			slog.String("error", err.Error()),
		)

		return err
	}

	// スタッフだけがbotを用いてdmを送信できるようにする
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

	// スタッフじゃなければはじく
	if !user.IsStaff {
		slog.WarnContext(
			e.Request().Context(),
			"user is not a staff member",
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	// 指定時刻まで待機してからDMを送信する
	go func() {
		if !req.SendAt.IsZero() {
			time.Sleep(time.Until(req.SendAt))
		}

		err := s.traqService.PostDirectMessage(context.Background(), string(userID), req.Content)
		if err != nil {
			slog.ErrorContext(
				context.Background(),
				"failed to send direct message",
				slog.String("error", err.Error()),
				slog.String("userId", string(userID)),
			)
		}
	}()

	return e.NoContent(http.StatusAccepted)
}

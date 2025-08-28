package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
)

// AdminPostMessage は DM を送信するハンドラです。
func (s *Server) AdminPostMessage(
	e echo.Context,
	targetUserID api.UserId,
	params api.AdminPostMessageParams,
) error {
	var req api.AdminPostMessageJSONRequestBody
	if err := e.Bind(&req); err != nil {
		return err
	}

	// スタッフだけがbotを用いてdmを送信できるようにする
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
	}

	// スタッフじゃなければはじく
	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	message, err := converter.Convert[model.Message](req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert request body: %w", err))
	}

	message.TargetUserID = targetUserID

	if err := s.repo.CreateMessage(e.Request().Context(), &message); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "User not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to create message: %w", err))
	}

	return e.NoContent(http.StatusAccepted)
}

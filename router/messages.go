package router

import (
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	traq "github.com/traPtitech/go-traq"

	"github.com/traPtitech/rucQ/api"
)

// AdminPostMessage は DM を送信するハンドラです。
func (s *Server) AdminPostMessage(
	e echo.Context,
	_ api.UserId,
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
			sendTime := req.SendAt
			duration := time.Until(sendTime)
			if duration > 0 {
				time.Sleep(duration) // 指定時刻まで待機
			}
		}
		postMessageRequest := *traq.NewPostMessageRequest(req.Content)
		postMessageRequest.SetEmbed(true)
		// targetUser, err := s.repo.GetOrCreateUser(string(userId))
		// if err != nil {
		// 	e.Logger().Errorf("failed to get or create user: %v", err)
		// 	return
		// }
		configuration := traq.NewConfiguration()
		configuration.AddDefaultHeader("Authorization", "Bearer "+os.Getenv("BOT_ACCESS_TOKEN"))
		// apiClient := traq.NewAPIClient(configuration)
		// _, r, err := apiClient.MessageApi.PostDirectMessage(context.Background(), targetUser.TraqUUID).PostMessageRequest(postMessageRequest).Execute()
		// if err != nil {
		// 	e.Logger().Errorf("error when calling MessageApi.PostDirectMessage: %v\nfull HTTP response: %v", err, r)
		// }
	}()

	return e.NoContent(http.StatusAccepted)
}

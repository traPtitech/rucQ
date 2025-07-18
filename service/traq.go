//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock

package service

import (
	"context"

	traq "github.com/traPtitech/go-traq"
)

// TraqService はtraQ APIとの連携を担当するサービスです。
type TraqService interface {
	// PostDirectMessage は指定したユーザーにダイレクトメッセージを送信します。
	PostDirectMessage(ctx context.Context, userID string, content string) error
}

type traqServiceImpl struct {
	client      *traq.APIClient
	accessToken string
}

// NewTraqService はTraqServiceを生成します。
func NewTraqService(baseURL, accessToken string) TraqService {
	config := traq.NewConfiguration()
	config.Servers = traq.ServerConfigurations{
		{
			URL: baseURL,
		},
	}
	client := traq.NewAPIClient(config)

	return &traqServiceImpl{
		client:      client,
		accessToken: accessToken,
	}
}

func (s *traqServiceImpl) PostDirectMessage(
	ctx context.Context,
	userID string,
	content string,
) error {
	authCtx := context.WithValue(ctx, traq.ContextAccessToken, s.accessToken)

	user, _, err := s.client.UserApi.GetUser(authCtx, userID).Execute()
	if err != nil {
		return err
	}

	postMessageRequest := *traq.NewPostMessageRequest(content)
	postMessageRequest.SetEmbed(true)

	req := s.client.MessageApi.PostDirectMessage(authCtx, user.Id).
		PostMessageRequest(postMessageRequest)

	_, _, err = req.Execute()
	return err
}

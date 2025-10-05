package traq

import (
	"context"

	traq "github.com/traPtitech/go-traq"
)

type traqServiceImpl struct {
	client      *traq.APIClient
	accessToken string
}

func NewTraqService(baseURL, accessToken string) *traqServiceImpl {
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

func (s *traqServiceImpl) GetCanonicalUserName(
	ctx context.Context,
	userID string,
) (string, error) {
	authCtx := context.WithValue(ctx, traq.ContextAccessToken, s.accessToken)
	users, _, err := s.client.UserApi.GetUsers(authCtx).Name(userID).Execute()

	if err != nil {
		return "", err
	}

	if len(users) == 0 {
		return "", ErrUserNotFound
	}

	return users[0].Name, nil
}

func (s *traqServiceImpl) PostDirectMessage(
	ctx context.Context,
	userID string,
	content string,
) error {
	authCtx := context.WithValue(ctx, traq.ContextAccessToken, s.accessToken)

	users, _, err := s.client.UserApi.GetUsers(authCtx).Name(userID).Execute()

	if err != nil {
		return err
	}

	if len(users) == 0 {
		return ErrUserNotFound
	}

	userUUID := users[0].Id
	postMessageRequest := *traq.NewPostMessageRequest(content)
	postMessageRequest.SetEmbed(true)

	req := s.client.MessageApi.PostDirectMessage(authCtx, userUUID).
		PostMessageRequest(postMessageRequest)

	_, _, err = req.Execute()
	return err
}

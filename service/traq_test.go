package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/traPtitech/rucQ/testutil/bot"
)

// setupTraqContainer starts MariaDB and traQ containers using compose and returns the traQ URL and access token
func setupTraqContainer(t *testing.T) (string, string) {
	t.Helper()
	ctx := context.Background()

	composeStack, err := compose.NewDockerComposeWith(
		compose.WithStackFiles("../compose.yaml"),
	)
	require.NoError(t, err, "Failed to create compose stack")

	// Set random ports via environment variables
	composeWithEnv := composeStack.WithEnv(map[string]string{
		"MARIADB_PORT":     "0",
		"TRAQ_SERVER_PORT": "0",
	})

	t.Cleanup(func() {
		require.NoError(
			t,
			composeStack.Down(
				ctx,
				compose.RemoveOrphans(true),
				compose.RemoveImagesLocal,
				compose.RemoveVolumes(true),
			),
		)
	})

	// Configure wait strategies for required services and start only those services
	composeWithWait := composeWithEnv.
		WaitForService("mariadb", wait.ForHealthCheck().WithStartupTimeout(60*time.Second)).
		WaitForService("traq_server", wait.ForHTTP("/api/v3/version").WithPort("3000/tcp").WithStartupTimeout(120*time.Second))

	err = composeWithWait.Up(ctx, compose.Wait(true), compose.RunServices("mariadb", "traq_server"))
	require.NoError(t, err, "Failed to start compose stack")

	// Get traQ server container
	traqContainer, err := composeStack.ServiceContainer(ctx, "traq_server")
	require.NoError(t, err, "Failed to get traQ server container")

	traqHost, err := traqContainer.Host(ctx)
	require.NoError(t, err)
	traqPort, err := traqContainer.MappedPort(ctx, "3000")
	require.NoError(t, err)

	traqURL := fmt.Sprintf("http://%s:%s", traqHost, traqPort.Port())
	accessToken, err := bot.CreateBot(traqURL)

	if err != nil {
		t.Fatalf("Failed to create test bot: %v", err)
	}

	return traqURL, accessToken
}

func TestTraqService_PostDirectMessage(t *testing.T) {
	traqURL, accessToken := setupTraqContainer(t)

	ctx := context.Background()
	service := NewTraqService(traqURL, accessToken)

	t.Run("存在しないユーザーへのメッセージ送信はエラーになる", func(t *testing.T) {
		err := service.PostDirectMessage(ctx, "nonexistent-user", "Test message")
		if err == nil {
			t.Error("Expected error for nonexistent user, but got nil")
		}
		t.Logf("Expected error occurred: %v", err)
	})

	t.Run("不正なアクセストークンでエラーになる", func(t *testing.T) {
		invalidService := NewTraqService(traqURL, "invalid-token")
		err := invalidService.PostDirectMessage(ctx, "traq", "Test message")
		if err == nil {
			t.Error("Expected error for invalid token, but got nil")
		}
		t.Logf("Expected error occurred: %v", err)
	})
}

package service

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/traPtitech/rucQ/testutil/bot"
	"github.com/traPtitech/rucQ/testutil/random"
)

func setup(t *testing.T) *traqServiceImpl {
	t.Helper()

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
				context.Background(),
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

	err = composeWithWait.Up(
		t.Context(),
		compose.Wait(true),
		compose.RunServices("mariadb", "traq_server"),
	)
	require.NoError(t, err, "Failed to start compose stack")

	// Get traQ server container
	traqContainer, err := composeStack.ServiceContainer(t.Context(), "traq_server")
	require.NoError(t, err, "Failed to get traQ server container")

	traqHost, err := traqContainer.Host(t.Context())
	require.NoError(t, err)
	traqPort, err := traqContainer.MappedPort(t.Context(), "3000")
	require.NoError(t, err)

	traqAPIBaseURL := fmt.Sprintf("http://%s:%s/api/v3", traqHost, traqPort.Port())
	accessToken, err := bot.CreateBot(traqAPIBaseURL)

	require.NoError(t, err, "Failed to create bot")

	return NewTraqService(traqAPIBaseURL, accessToken)
}

const existingUserID = "traq" // 既存のユーザーID

func TestTraqServiceImpl_GetCanonicalUserName(t *testing.T) {
	t.Parallel()

	t.Run("存在するユーザーの正規化された名前を取得できる", func(t *testing.T) {
		t.Parallel()

		s := setup(t)
		userName, err := s.GetCanonicalUserName(t.Context(), strings.ToUpper(existingUserID))

		assert.NoError(t, err)
		assert.Equal(t, existingUserID, userName)
	})

	t.Run("存在しないユーザーの場合はErrUserNotFoundを返す", func(t *testing.T) {
		t.Parallel()

		s := setup(t)
		userID := random.AlphaNumericString(t, 32) // 非存在ユーザーID
		_, err := s.GetCanonicalUserName(t.Context(), userID)

		if assert.Error(t, err, "Expected error for nonexistent user, but got nil") {
			assert.Equal(t, ErrUserNotFound, err)
		}
	})
}

func TestTraqServiceImpl_PostDirectMessage(t *testing.T) {
	t.Parallel()

	t.Run("存在するユーザーへのメッセージ送信は成功する", func(t *testing.T) {
		t.Parallel()

		s := setup(t)
		message := random.AlphaNumericString(t, 100)
		err := s.PostDirectMessage(t.Context(), existingUserID, message)

		assert.NoError(t, err)
	})

	t.Run("存在しないユーザーへのメッセージ送信はエラーになる", func(t *testing.T) {
		t.Parallel()

		s := setup(t)
		userID := random.AlphaNumericString(t, 32) // 非存在ユーザーID
		message := random.AlphaNumericString(t, 100)
		err := s.PostDirectMessage(t.Context(), userID, message)

		if assert.Error(t, err, "Expected error for nonexistent user, but got nil") {
			assert.Equal(t, ErrUserNotFound, err)
		}
	})
}

package service

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/traPtitech/rucQ/testutil/bot"
	"github.com/traPtitech/rucQ/testutil/random"
)

var s *traqServiceImpl

func TestMain(m *testing.M) {
	composeStack, err := compose.NewDockerComposeWith(
		compose.WithStackFiles("../compose.yaml"),
	)

	if err != nil {
		panic(fmt.Sprintf("failed to create compose stack: %v", err))
	}

	defer func() {
		composeStack.Down(
			context.Background(),
			compose.RemoveOrphans(true),
			compose.RemoveImagesLocal,
			compose.RemoveVolumes(true),
		)
	}()

	ctx := context.Background()

	if err := composeStack.
		WithEnv(map[string]string{
			"MARIADB_PORT":     "0",
			"TRAQ_SERVER_PORT": "0",
		}).
		WaitForService(
			"mariadb",
			wait.ForHealthCheck().WithStartupTimeout(60*time.Second),
		).
		WaitForService(
			"traq_server",
			wait.ForHTTP("/api/v3/version").WithPort("3000/tcp").WithStartupTimeout(120*time.Second),
		).
		Up(
			ctx,
			compose.Wait(true),
			compose.RunServices("mariadb", "traq_server"),
		); err != nil {
		panic(fmt.Sprintf("failed to start compose stack: %v", err))
	}

	traqContainer, err := composeStack.ServiceContainer(ctx, "traq_server")

	if err != nil {
		panic(fmt.Sprintf("failed to get traQ server container: %v", err))
	}

	traqHost, err := traqContainer.Host(ctx)

	if err != nil {
		panic(fmt.Sprintf("failed to get host: %v", err))
	}

	traqPort, err := traqContainer.MappedPort(ctx, "3000")

	if err != nil {
		panic(fmt.Sprintf("failed to get mapped port: %v", err))
	}

	traqAPIBaseURL := fmt.Sprintf("http://%s:%s/api/v3", traqHost, traqPort.Port())
	accessToken, err := bot.CreateBot(traqAPIBaseURL)

	if err != nil {
		panic(fmt.Sprintf("failed to create bot: %v", err))
	}

	s = NewTraqService(traqAPIBaseURL, accessToken)

	m.Run()
}

const existingUserID = "traq" // 既存のユーザーID

func TestTraqServiceImpl_GetCanonicalUserName(t *testing.T) {
	t.Parallel()

	t.Run("存在するユーザーの正規化された名前を取得できる", func(t *testing.T) {
		t.Parallel()

		userName, err := s.GetCanonicalUserName(t.Context(), strings.ToUpper(existingUserID))

		assert.NoError(t, err)
		assert.Equal(t, existingUserID, userName)
	})

	t.Run("存在しないユーザーの場合はErrUserNotFoundを返す", func(t *testing.T) {
		t.Parallel()

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

		message := random.AlphaNumericString(t, 100)
		err := s.PostDirectMessage(t.Context(), existingUserID, message)

		assert.NoError(t, err)
	})

	t.Run("存在しないユーザーへのメッセージ送信はエラーになる", func(t *testing.T) {
		t.Parallel()

		userID := random.AlphaNumericString(t, 32) // 非存在ユーザーID
		message := random.AlphaNumericString(t, 100)
		err := s.PostDirectMessage(t.Context(), userID, message)

		if assert.Error(t, err, "Expected error for nonexistent user, but got nil") {
			assert.Equal(t, ErrUserNotFound, err)
		}
	})
}

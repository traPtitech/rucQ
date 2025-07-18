package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setup はテストごとに独立したtraQ環境をセットアップします。
func setup(t *testing.T) TraqService {
	t.Helper()

	ctx := context.Background()

	// MariaDBコンテナを起動
	mariadbReq := testcontainers.ContainerRequest{
		Image:        "mariadb:11.8.2-noble",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MARIADB_ROOT_PASSWORD": "password",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("ready for connections").WithOccurrence(2), // 初期化と本起動の2回
			wait.ForSQL("3306/tcp", "mysql", func(host string, port nat.Port) string {
				return fmt.Sprintf("root:password@tcp(%s:%s)/", host, port.Port())
			}),
		),
	}
	mariadbContainer, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: mariadbReq,
			Started:          true,
		},
	)
	require.NoError(t, err, "failed to start mariadb container")

	t.Cleanup(func() {
		testcontainers.CleanupContainer(t, mariadbContainer)
	})

	mariadbPort, err := mariadbContainer.MappedPort(ctx, "3306")
	require.NoError(t, err, "failed to get mapped port for mariadb")

	// traQ serverコンテナを起動
	traqReq := testcontainers.ContainerRequest{
		Image:        "ghcr.io/traptitech/traq:3.24.13",
		ExposedPorts: []string{"3000/tcp"},
		Env: map[string]string{
			"TRAQ_MARIADB_HOST":      "host.docker.internal",
			"TRAQ_MARIADB_PORT":      mariadbPort.Port(),
			"TRAQ_MARIADB_USERNAME":  "root",
			"TRAQ_MARIADB_PASSWORD":  "password",
			"TRAQ_MARIADB_DATABASE":  "traq",
			"TRAQ_STORAGE_LOCAL_DIR": "/app/storage",
		},
		WaitingFor: wait.ForHTTP("/api/v3/version").
			WithPort("3000/tcp").
			WithStartupTimeout(5 * time.Minute),
	}
	traqContainer, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: traqReq,
			Started:          true,
		},
	)
	require.NoError(t, err, "failed to start traq container")

	t.Cleanup(func() {
		testcontainers.CleanupContainer(t, traqContainer)
	})

	traqPort, err := traqContainer.MappedPort(ctx, "3000")
	require.NoError(t, err, "failed to get mapped port for traq")
	traqBaseURL := fmt.Sprintf("http://localhost:%s", traqPort.Port())
	traqApiURL := traqBaseURL + "/api/v3"

	// ログインしてセッションクッキーを取得
	loginBody := map[string]string{
		"name":     "traq",
		"password": "traq",
	}
	loginBodyBytes, err := json.Marshal(loginBody)
	require.NoError(t, err)

	var loginRes *http.Response
	var setCookie string
	httpClient := &http.Client{}

	// ログインをリトライ（コンテナ起動直後はAPIが不安定な場合があるため）
	require.Eventually(t, func() bool {
		loginReq, err := http.NewRequest(
			"POST",
			traqApiURL+"/login",
			bytes.NewBuffer(loginBodyBytes),
		)
		if err != nil {
			return false
		}
		loginReq.Header.Set("Content-Type", "application/json")

		loginRes, err = httpClient.Do(loginReq)
		if err != nil {
			return false
		}
		defer loginRes.Body.Close()

		if loginRes.StatusCode == http.StatusNoContent {
			for _, cookie := range loginRes.Cookies() {
				if cookie.Name != "" {
					setCookie = cookie.Name + "=" + cookie.Value
					return true
				}
			}
		}
		return false
	}, 30*time.Second, 1*time.Second, "failed to login as traq user")

	require.NotEmpty(t, setCookie, "no session cookie found")

	// テスト用Botの作成
	botName := fmt.Sprintf("test-bot-%s", uuid.New().String()[:8])
	botBody := map[string]interface{}{
		"name":        botName,
		"displayName": "Test Bot",
		"description": "Test Bot for integration test",
		"mode":        "HTTP",
		"endpoint":    "",
		"scopes":      []string{"read", "write"},
	}
	botBodyBytes, err := json.Marshal(botBody)
	require.NoError(t, err)

	botReq, err := http.NewRequest("POST", traqApiURL+"/bots", bytes.NewBuffer(botBodyBytes))
	require.NoError(t, err)
	botReq.Header.Set("Content-Type", "application/json")
	botReq.Header.Set("Cookie", setCookie)

	botRes, err := httpClient.Do(botReq)
	require.NoError(t, err)
	defer botRes.Body.Close()

	require.Equal(t, http.StatusCreated, botRes.StatusCode, "failed to create bot")

	var botResponse struct {
		Tokens struct {
			AccessToken string `json:"access_token"`
		} `json:"tokens"`
	}
	err = json.NewDecoder(botRes.Body).Decode(&botResponse)
	require.NoError(t, err)
	require.NotEmpty(t, botResponse.Tokens.AccessToken, "bot access token is empty")

	return NewTraqService(traqApiURL, botResponse.Tokens.AccessToken)
}

func TestTraqServiceImpl_PostDirectMessage(t *testing.T) {
	t.Parallel()
	s := setup(t)
	require.NotNil(t, s)

	const content = "これはtestcontainersからのメッセージです。"

	t.Run("正常系: DMを送信できる", func(t *testing.T) {
		t.Parallel()
		// traQのデフォルトで存在するユーザー "traq" を使用
		const targetUserID = "traq"
		err := s.PostDirectMessage(context.Background(), targetUserID, content)
		assert.NoError(t, err)
	})

	t.Run("異常系: 存在しないユーザー", func(t *testing.T) {
		t.Parallel()
		const nonExistentUserID = "a-z-a-z-a-z-a-z-a-z-a-z-a-z"
		err := s.PostDirectMessage(context.Background(), nonExistentUserID, content)
		assert.Error(t, err, "存在しないユーザーへの送信はエラーになるはずです")
	})
}

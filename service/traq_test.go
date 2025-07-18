package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupTraqContainer starts MariaDB and traQ containers and returns the traQ URL and access token
func setupTraqContainer(t *testing.T) (string, string, func()) {
	ctx := context.Background()

	// Start MariaDB container
	mariadbContainer, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "mariadb:11.8.2-noble",
				ExposedPorts: []string{"3306/tcp"},
				Env: map[string]string{
					"MARIADB_ROOT_PASSWORD": "password",
					"MARIADB_DATABASE":      "traq",
				},
				WaitingFor: wait.ForAll(
					wait.ForLog("ready for connections"),
					wait.ForListeningPort("3306/tcp"),
				).WithStartupTimeout(60 * time.Second),
			},
			Started: true,
		},
	)
	if err != nil {
		t.Fatalf("Failed to start MariaDB container: %v", err)
	}

	// Get MariaDB container IP
	mariadbIP, err := mariadbContainer.ContainerIP(ctx)
	if err != nil {
		t.Fatalf("Failed to get MariaDB IP: %v", err)
	}

	// Wait a bit more to ensure MariaDB is fully ready
	time.Sleep(2 * time.Second)

	// Start traQ container
	traqContainer, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "ghcr.io/traptitech/traq:3.20.2",
				ExposedPorts: []string{"3000/tcp"},
				Env: map[string]string{
					"TRAQ_MARIADB_HOST":     mariadbIP,
					"TRAQ_MARIADB_PORT":     "3306",
					"TRAQ_MARIADB_USERNAME": "root",
					"TRAQ_MARIADB_PASSWORD": "password",
					"TRAQ_MARIADB_DATABASE": "traq",
					"TRAQ_ORIGIN":           "http://localhost:3000",
				},
				WaitingFor: wait.ForHTTP("/api/v3/version").
					WithPort("3000/tcp").
					WithStartupTimeout(120 * time.Second),
			},
			Started: true,
		},
	)
	if err != nil {
		t.Fatalf("Failed to start traQ container: %v", err)
	}

	// Get traQ container port
	traqPort, err := traqContainer.MappedPort(ctx, "3000")
	if err != nil {
		t.Fatalf("Failed to get traQ port: %v", err)
	}

	traqURL := fmt.Sprintf("http://localhost:%s", traqPort.Port())

	// Create bot and get access token
	accessToken := createTestBot(t, traqURL)

	cleanup := func() {
		traqContainer.Terminate(ctx)
		mariadbContainer.Terminate(ctx)
	}

	return traqURL, accessToken, cleanup
}

// createTestBot creates a test bot and returns its access token
func createTestBot(t *testing.T, traqURL string) string {
	// Login to get session cookie
	loginPayload := map[string]string{
		"name":     "traq",
		"password": "traq",
	}
	loginBytes, _ := json.Marshal(loginPayload)

	loginResp, err := http.Post(
		traqURL+"/api/v3/login",
		"application/json",
		bytes.NewBuffer(loginBytes),
	)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	defer loginResp.Body.Close()

	// Extract session cookie
	var sessionCookie *http.Cookie
	for _, cookie := range loginResp.Cookies() {
		if cookie.Name == "r_session" {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatalf("No session cookie found")
	}

	// Create bot
	createBotPayload := map[string]interface{}{
		"name":        "test-bot",
		"displayName": "Test Bot",
		"description": "Test bot for integration tests",
		"mode":        "HTTP",
		"endpoint":    "http://localhost:3001/webhook",
	}
	payloadBytes, _ := json.Marshal(createBotPayload)

	req, err := http.NewRequest("POST", traqURL+"/api/v3/bots", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(sessionCookie)

	client := &http.Client{}
	createBotResp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}
	defer createBotResp.Body.Close()

	if createBotResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(createBotResp.Body)
		t.Fatalf(
			"Failed to create bot: status %d, body: %s",
			createBotResp.StatusCode,
			string(body),
		)
	}

	// Parse bot response to get access token
	var botResponse struct {
		VerificationToken string `json:"verificationToken"`
		AccessToken       string `json:"accessToken"`
		BotUserID         string `json:"botUserId"`
	}

	body, err := io.ReadAll(createBotResp.Body)
	if err != nil {
		t.Fatalf("Failed to read bot response: %v", err)
	}

	if err := json.Unmarshal(body, &botResponse); err != nil {
		t.Fatalf("Failed to parse bot response: %v", err)
	}

	return botResponse.AccessToken
}

func TestTraqService_PostDirectMessage(t *testing.T) {
	traqURL, accessToken, cleanup := setupTraqContainer(t)
	defer cleanup()

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

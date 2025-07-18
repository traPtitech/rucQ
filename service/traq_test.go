package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/traPtitech/rucQ/testutil/port"
)

// sanitizeStackIdentifier removes special characters from test names to create valid Docker Compose stack identifiers
func sanitizeStackIdentifier(testName string) string {
	// Replace all non-alphanumeric characters with hyphens (including underscores for Docker compatibility)
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	sanitized := re.ReplaceAllString(testName, "-")

	// Remove consecutive hyphens
	re = regexp.MustCompile(`-+`)
	sanitized = re.ReplaceAllString(sanitized, "-")

	// Trim leading/trailing hyphens and convert to lowercase
	sanitized = strings.Trim(sanitized, "-")
	sanitized = strings.ToLower(sanitized)

	// Ensure it starts with a letter, if not, prefix with "test"
	if len(sanitized) == 0 || !(sanitized[0] >= 'a' && sanitized[0] <= 'z') {
		sanitized = "test-" + sanitized
	}

	// Limit length to avoid overly long names
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}

	return sanitized
}

// setupTraqContainer starts MariaDB and traQ containers using compose and returns the traQ URL and access token
func setupTraqContainer(t *testing.T) (string, string) {
	t.Helper()
	ctx := context.Background()

	// Generate random ports to avoid conflicts between parallel tests
	portNames := []string{
		"MARIADB_PORT",
		"RUCQ_PORT",
		"SWAGGER_PORT",
		"ADMINER_PORT",
		"TRAQ_CADDY_PORT",
		"TRAQ_SERVER_PORT",
	}
	randomPorts := port.MustGetFreePorts(len(portNames))
	portEnvMap := port.PortsToStringMap(portNames, randomPorts)

	// Create a compose stack using the root compose.yaml file with a unique stack identifier
	// This ensures each test gets its own set of containers with unique names/networks
	stackIdentifier := sanitizeStackIdentifier(fmt.Sprintf("test-%s-%d", t.Name(), rand.Int()))
	composeStack, err := compose.NewDockerComposeWith(
		compose.WithStackFiles("../compose.yaml"),
		compose.StackIdentifier(stackIdentifier),
	)
	require.NoError(t, err, "Failed to create compose stack")

	// Set random ports via environment variables
	composeWithEnv := composeStack.WithEnv(portEnvMap)

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

	// Configure wait strategies for required services and start all services
	composeWithWait := composeWithEnv.
		WaitForService("mariadb", wait.ForHealthCheck().WithStartupTimeout(60*time.Second)).
		WaitForService("traq_server", wait.ForHTTP("/api/v3/version").WithPort("3000/tcp").WithStartupTimeout(120*time.Second))

	err = composeWithWait.Up(ctx, compose.Wait(true))
	require.NoError(t, err, "Failed to start compose stack")

	// Stop unnecessary services to avoid port conflicts with other tests
	stopServices := []string{"rucq", "swagger", "adminer", "traq_caddy", "traq_ui"}
	for _, service := range stopServices {
		// Get service container and stop it (ignore errors if service doesn't exist or isn't running)
		if container, err := composeStack.ServiceContainer(ctx, service); err == nil {
			_ = container.Stop(ctx, nil)
		}
	}

	// Get traQ server container
	traqContainer, err := composeStack.ServiceContainer(ctx, "traq_server")
	require.NoError(t, err, "Failed to get traQ server container")

	traqHost, err := traqContainer.Host(ctx)
	require.NoError(t, err)
	traqPort, err := traqContainer.MappedPort(ctx, "3000")
	require.NoError(t, err)

	traqURL := fmt.Sprintf("http://%s:%s", traqHost, traqPort.Port())

	// Create test bot and get access token
	accessToken := createTestBot(t, traqURL)

	return traqURL, accessToken
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

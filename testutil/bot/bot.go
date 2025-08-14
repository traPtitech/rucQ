package bot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

// CreateBot はtraQにログインし、"rucq"という名前のBotを作成してアクセストークンを返します。
// ログイン情報は "traq"/"traq" に、Bot設定は "rucq" のものに固定されています。
func CreateBot(traqAPIBaseURL string) (string, error) {
	client := &http.Client{}

	// 1. ログインしてセッションクッキーを取得
	sessionCookie, err := login(client, traqAPIBaseURL, "traq", "traq")
	if err != nil {
		return "", fmt.Errorf("login failed: %w", err)
	}

	// 2. Botを作成 (設定は固定)
	botConfig := map[string]string{
		"name":        "rucq",
		"displayName": "rucQ",
		"description": "rucQ Bot",
		"mode":        "HTTP",
		"endpoint":    "http://example.com",
	}
	payloadBytes, err := json.Marshal(botConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal create bot payload: %w", err)
	}

	botEndpoint, err := url.JoinPath(traqAPIBaseURL, "bots")

	if err != nil {
		return "", fmt.Errorf("failed to construct bot creation endpoint URL: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		botEndpoint,
		bytes.NewBuffer(payloadBytes),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request for bot creation: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(sessionCookie)

	createBotResp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request to create bot: %w", err)
	}
	defer func() {
		err := createBotResp.Body.Close()

		if err != nil {
			log.Fatal("failed to close response body:", err)
		}
	}()

	body, err := io.ReadAll(createBotResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read bot creation response body: %w", err)
	}

	if createBotResp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf(
			"failed to create bot: status %d, body: %s",
			createBotResp.StatusCode,
			string(body),
		)
	}

	// 3. レスポンスをパースしてアクセストークンを取得
	var botResponse struct {
		Tokens struct {
			AccessToken string `json:"accessToken"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(body, &botResponse); err != nil {
		return "", fmt.Errorf("failed to parse bot response: %w. body: %s", err, string(body))
	}

	if botResponse.Tokens.AccessToken == "" {
		return "", fmt.Errorf("access token not found in bot response: %s", string(body))
	}

	return botResponse.Tokens.AccessToken, nil
}

func login(client *http.Client, traqAPIBaseURL, username, password string) (*http.Cookie, error) {
	loginPayload := map[string]string{
		"name":     username,
		"password": password,
	}
	loginBytes, err := json.Marshal(loginPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login payload: %w", err)
	}

	loginEndpoint, err := url.JoinPath(traqAPIBaseURL, "login")

	if err != nil {
		return nil, fmt.Errorf("failed to construct login endpoint URL: %w", err)
	}

	loginResp, err := client.Post(
		loginEndpoint,
		"application/json",
		bytes.NewBuffer(loginBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute login request: %w", err)
	}
	defer func() {
		err := loginResp.Body.Close()

		if err != nil {
			log.Fatal("failed to close response body:", err)
		}
	}()

	if loginResp.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(loginResp.Body)
		if err != nil {
			return nil, fmt.Errorf(
				"login request failed: status %d, failed to read response body: %w",
				loginResp.StatusCode,
				err,
			)
		}
		return nil, fmt.Errorf(
			"login request failed: status %d, body: %s",
			loginResp.StatusCode,
			string(body),
		)
	}

	for _, cookie := range loginResp.Cookies() {
		if cookie.Name == "r_session" {
			return cookie, nil
		}
	}

	return nil, errors.New("session cookie not found in login response")
}

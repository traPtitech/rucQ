package bot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// CreateBot はtraQにログインし、"rucq"という名前のBotを作成してアクセストークンを返します。
// ログイン情報は "traq"/"traq" に、Bot設定は "rucq" のものに固定されています。
func CreateBot(traqURL string) (string, error) {
	client := &http.Client{}

	// 1. ログインしてセッションクッキーを取得
	sessionCookie, err := login(client, traqURL, "traq", "traq")
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

	req, err := http.NewRequest("POST", traqURL+"/api/v3/bots", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request for bot creation: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(sessionCookie)

	createBotResp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request to create bot: %w", err)
	}
	defer createBotResp.Body.Close()

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

func login(client *http.Client, traqURL, username, password string) (*http.Cookie, error) {
	loginPayload := map[string]string{
		"name":     username,
		"password": password,
	}
	loginBytes, err := json.Marshal(loginPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login payload: %w", err)
	}

	loginResp, err := client.Post(
		traqURL+"/api/v3/login",
		"application/json",
		bytes.NewBuffer(loginBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute login request: %w", err)
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(loginResp.Body)
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

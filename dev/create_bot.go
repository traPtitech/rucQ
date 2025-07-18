package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	const TRAQ_BASE_URL = "http://localhost:3000/api/v3"

	// ログイン
	loginBody := map[string]string{
		"name":     "traq",
		"password": "traq",
	}
	loginBodyBytes, _ := json.Marshal(loginBody)
	loginReq, _ := http.NewRequest("POST", TRAQ_BASE_URL+"/login", bytes.NewBuffer(loginBodyBytes))
	loginReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	loginRes, err := client.Do(loginReq)
	if err != nil {
		fmt.Println("Login failed:", err)
		os.Exit(1)
	}
	defer loginRes.Body.Close()

	if loginRes.StatusCode < 200 || loginRes.StatusCode >= 300 {
		body, _ := io.ReadAll(loginRes.Body)
		fmt.Printf("HTTP error! status: %d, response: %s\n", loginRes.StatusCode, string(body))
		os.Exit(1)
	}

	setCookie := ""
	for _, cookie := range loginRes.Cookies() {
		if cookie.Name != "" {
			setCookie = cookie.Name + "=" + cookie.Value
			break
		}
	}
	if setCookie == "" {
		fmt.Println("No session cookie found")
		os.Exit(1)
	}

	botName := "rucq"
	botDisplayName := "rucQ"

	botBody := map[string]string{
		"name":        botName,
		"displayName": botDisplayName,
		"description": "rucQ Bot",
		"mode":        "HTTP",
		"endpoint":    "http://example.com",
	}
	botBodyBytes, _ := json.Marshal(botBody)
	botReq, _ := http.NewRequest("POST", TRAQ_BASE_URL+"/bots", bytes.NewBuffer(botBodyBytes))
	botReq.Header.Set("Content-Type", "application/json")
	botReq.Header.Set("Cookie", setCookie)

	botRes, err := client.Do(botReq)
	if err != nil {
		fmt.Println("Failed to create bot:", err)
		os.Exit(1)
	}
	defer botRes.Body.Close()

	if botRes.StatusCode < 200 || botRes.StatusCode >= 300 {
		body, _ := io.ReadAll(botRes.Body)
		fmt.Printf("HTTP error! status: %d, response: %s\n", botRes.StatusCode, string(body))
		os.Exit(1)
	}

	var responseData map[string]any
	json.NewDecoder(botRes.Body).Decode(&responseData)

	// アクセストークン出力
	tokens, ok := responseData["tokens"].(map[string]any)
	if ok {
		accessToken, ok := tokens["accessToken"].(string)
		if ok {
			fmt.Println("TRAQ_BOT_TOKEN=" + accessToken)
		}
	}
}

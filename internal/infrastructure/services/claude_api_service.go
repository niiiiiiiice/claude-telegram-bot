package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"telegram-chatbot/internal/domain/entities"
	"telegram-chatbot/internal/domain/services"
	"time"
)

type ClaudeAPIService struct {
	apiKey     string
	httpClient *http.Client
}

func NewClaudeAPIService(apiKey string) services.ClaudeService {
	return &ClaudeAPIService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
	System    string          `json:"system,omitempty"`
}

type ClaudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func (s *ClaudeAPIService) GenerateResponse(messages []entities.Message) (string, error) {
	claudeMessages := make([]ClaudeMessage, 0, len(messages))

	for _, msg := range messages {
		claudeMessages = append(claudeMessages, ClaudeMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	request := ClaudeRequest{
		Model:     "claude-3-5-sonnet-20241022", // Экономичная модель с доступом в интернет
		MaxTokens: 1024,
		Messages:  claudeMessages,
		System:    "Ты семейный помощник-бот. Отвечай дружелюбно и полезно на русском языке.",
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude API")
	}

	return claudeResp.Content[0].Text, nil
}

package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	TelegramBotToken string
	ClaudeAPIKey     string
	AllowedChatID    int64
	LogLevel         string
}

func Load() (*Config, error) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	claudeAPIKey := os.Getenv("CLAUDE_API_KEY")
	if claudeAPIKey == "" {
		return nil, fmt.Errorf("CLAUDE_API_KEY is required")
	}

	chatIDStr := os.Getenv("ALLOWED_CHAT_ID")
	if chatIDStr == "" {
		return nil, fmt.Errorf("ALLOWED_CHAT_ID is required")
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid ALLOWED_CHAT_ID: %v", err)
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return &Config{
		TelegramBotToken: botToken,
		ClaudeAPIKey:     claudeAPIKey,
		AllowedChatID:    chatID,
		LogLevel:         logLevel,
	}, nil
}

package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	TelegramBotToken string
	ClaudeAPIKey     string
	AllowedChatIDs   []int64
	LogLevel         string
	RedisHost        string
	RedisPort        string
	RedisUsername    string
	RedisPassword    string
	RedisDB          int
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

	chatIDsStr := os.Getenv("ALLOWED_CHAT_IDS")
	if chatIDsStr == "" {
		return nil, fmt.Errorf("ALLOWED_CHAT_IDS is required")
	}

	chatIDStrings := strings.Split(chatIDsStr, ",")
	chatIDs := make([]int64, 0, len(chatIDStrings))

	for _, idStr := range chatIDStrings {
		idStr = strings.TrimSpace(idStr)
		chatID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid chat ID in ALLOWED_CHAT_IDS: %s - %v", idStr, err)
		}
		chatIDs = append(chatIDs, chatID)
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	// Redis configuration
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redisUsername := os.Getenv("REDIS_USERNAME")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisDBStr := os.Getenv("REDIS_DB")
	redisDB := 0
	if redisDBStr != "" {
		var dbErr error
		redisDB, dbErr = strconv.Atoi(redisDBStr)
		if dbErr != nil {
			return nil, fmt.Errorf("invalid REDIS_DB: %v", dbErr)
		}
	}

	return &Config{
		TelegramBotToken: botToken,
		ClaudeAPIKey:     claudeAPIKey,
		AllowedChatIDs:   chatIDs,
		LogLevel:         logLevel,
		RedisHost:        redisHost,
		RedisPort:        redisPort,
		RedisUsername:    redisUsername,
		RedisPassword:    redisPassword,
		RedisDB:          redisDB,
	}, nil
}

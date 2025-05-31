package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"telegram-chatbot/internal/config"
	"telegram-chatbot/internal/domain/entities"
	"telegram-chatbot/internal/domain/repositories"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisSessionRepository struct {
	client *redis.Client
}

func NewRedisSessionRepository(cfg *config.Config) repositories.SessionRepository {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Username: cfg.RedisUsername,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	return &RedisSessionRepository{
		client: client,
	}
}

func (r *RedisSessionRepository) getKey(chatID, userID int64) string {
	return fmt.Sprintf("session:%d:%d", chatID, userID)
}

func (r *RedisSessionRepository) GetSession(chatID, userID int64) (*entities.ChatSession, error) {
	ctx := context.Background()
	key := r.getKey(chatID, userID)

	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		// Session doesn't exist, create a new one
		return &entities.ChatSession{
			ChatID:    chatID,
			UserID:    userID,
			IsActive:  false,
			Messages:  []entities.Message{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get session from Redis: %w", err)
	}

	var session entities.ChatSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &session, nil
}

func (r *RedisSessionRepository) SaveSession(session *entities.ChatSession) error {
	ctx := context.Background()
	key := r.getKey(session.ChatID, session.UserID)

	// Update the timestamp
	session.UpdatedAt = time.Now()

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Set with an expiration time (e.g., 24 hours)
	if err := r.client.Set(ctx, key, data, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to save session to Redis: %w", err)
	}

	return nil
}

func (r *RedisSessionRepository) DeleteSession(chatID, userID int64) error {
	ctx := context.Background()
	key := r.getKey(chatID, userID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session from Redis: %w", err)
	}

	return nil
}

func (r *RedisSessionRepository) IsSessionActive(chatID, userID int64) bool {
	session, err := r.GetSession(chatID, userID)
	if err != nil {
		return false
	}

	return session.IsActive
}

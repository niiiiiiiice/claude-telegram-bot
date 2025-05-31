package repositories

import (
	"telegram-chatbot/internal/domain/entities"
)

type SessionRepository interface {
	GetSession(chatID, userID int64) (*entities.ChatSession, error)
	SaveSession(session *entities.ChatSession) error
	DeleteSession(chatID, userID int64) error
	IsSessionActive(chatID, userID int64) bool
}

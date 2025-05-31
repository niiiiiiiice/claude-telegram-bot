package repositories

import (
	"fmt"
	"sync"
	"telegram-chatbot/internal/domain/entities"
	"telegram-chatbot/internal/domain/repositories"
	"time"
)

type MemorySessionRepository struct {
	sessions map[string]*entities.ChatSession
	mutex    sync.RWMutex
}

func NewMemorySessionRepository() repositories.SessionRepository {
	return &MemorySessionRepository{
		sessions: make(map[string]*entities.ChatSession),
	}
}

func (r *MemorySessionRepository) getKey(chatID, userID int64) string {
	return fmt.Sprintf("%d:%d", chatID, userID)
}

func (r *MemorySessionRepository) GetSession(chatID, userID int64) (*entities.ChatSession, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	key := r.getKey(chatID, userID)
	session, exists := r.sessions[key]
	if !exists {
		return &entities.ChatSession{
			ChatID:    chatID,
			UserID:    userID,
			IsActive:  false,
			Messages:  []entities.Message{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	}

	return session, nil
}

func (r *MemorySessionRepository) SaveSession(session *entities.ChatSession) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	key := r.getKey(session.ChatID, session.UserID)
	r.sessions[key] = session
	return nil
}

func (r *MemorySessionRepository) DeleteSession(chatID, userID int64) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	key := r.getKey(chatID, userID)
	delete(r.sessions, key)
	return nil
}

func (r *MemorySessionRepository) IsSessionActive(chatID, userID int64) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	key := r.getKey(chatID, userID)
	session, exists := r.sessions[key]
	return exists && session.IsActive
}

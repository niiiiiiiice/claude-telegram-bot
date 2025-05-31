package entities

import (
	"time"
)

type ChatSession struct {
	ChatID    int64
	UserID    int64
	IsActive  bool
	Messages  []Message
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Message struct {
	Role      string // "user" or "assistant"
	Content   string
	Timestamp time.Time
}

func (s *ChatSession) AddMessage(role, content string) {
	s.Messages = append(s.Messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	s.UpdatedAt = time.Now()
}

func (s *ChatSession) GetContextSize() int {
	size := 0
	for _, msg := range s.Messages {
		size += len(msg.Content)
	}
	return size
}

func (s *ChatSession) Reset() {
	s.Messages = []Message{}
	s.UpdatedAt = time.Now()
}

package services

import (
	"telegram-chatbot/internal/domain/entities"
)

type ClaudeService interface {
	GenerateResponse(messages []entities.Message) (string, error)
}

package handlers

import (
	"context"
	"fmt"
	"telegram-chatbot/internal/domain/commands"
	"telegram-chatbot/internal/domain/repositories"
	"telegram-chatbot/internal/domain/services"

	"go.uber.org/zap"
)

const MaxContextSize = 10000 // Примерный лимит символов

type CommandHandler struct {
	sessionRepo   repositories.SessionRepository
	claudeService services.ClaudeService
	logger        *zap.Logger
}

func NewCommandHandler(
	sessionRepo repositories.SessionRepository,
	claudeService services.ClaudeService,
	logger *zap.Logger,
) *CommandHandler {
	return &CommandHandler{
		sessionRepo:   sessionRepo,
		claudeService: claudeService,
		logger:        logger,
	}
}

func (h *CommandHandler) HandleStart(ctx context.Context, cmd commands.StartCommand) (string, error) {
	h.logger.Info("Handling start command", zap.Int64("chatID", cmd.ChatID), zap.Int64("userID", cmd.UserID))

	// Полный перезапуск - удаляем сессию
	if err := h.sessionRepo.DeleteSession(cmd.ChatID, cmd.UserID); err != nil {
		h.logger.Error("Failed to delete session", zap.Error(err))
		return "", err
	}

	return h.getHelpMessage(), nil
}

func (h *CommandHandler) HandleHelp(ctx context.Context, cmd commands.HelpCommand) (string, error) {
	h.logger.Info("Handling help command", zap.Int64("chatID", cmd.ChatID), zap.Int64("userID", cmd.UserID))
	return h.getHelpMessage(), nil
}

func (h *CommandHandler) getHelpMessage() string {
	return `🤖 **Семейный помощник-бот**

📋 **Доступные команды:**

/start - Перезапустить бота и показать это меню
/help - Показать справку по командам
/begin_chat - Начать сессию общения (бот запомнит контекст)
/end_chat - Завершить сессию и очистить контекст
/whoami - Показать информацию о пользователе и группе

💬 **Как использовать:**
• В группах упоминай меня @botname чтобы я ответил
• В личных сообщениях просто пиши - отвечу на всё
• Сессия позволяет мне помнить контекст разговора
• Если контекст станет слишком большим, я его автоматически очищу

✨ **Возможности:**
• Отвечаю на вопросы на русском языке
• Помогаю с задачами и советами
• Поддерживаю контекст во время активной сессии
• Работаю только в этой семейной группе

Начни с команды /begin_chat чтобы я запомнил наш разговор! 🚀`
}

func (h *CommandHandler) HandleBeginChat(ctx context.Context, cmd commands.StartBeginCommand) (string, error) {
	h.logger.Info("Handling start chat command", zap.Int64("chatID", cmd.ChatID), zap.Int64("userID", cmd.UserID))

	session, err := h.sessionRepo.GetSession(cmd.ChatID, cmd.UserID)
	if err != nil {
		return "", err
	}

	if session.IsActive {
		// Завершаем текущую сессию и начинаем новую
		session.Reset()
	}

	session.IsActive = true

	if err := h.sessionRepo.SaveSession(session); err != nil {
		return "", err
	}

	return "💬 Сессия общения начата! Теперь я буду запоминать контекст наших сообщений.", nil
}

func (h *CommandHandler) HandleEndChat(ctx context.Context, cmd commands.EndChatCommand) (string, error) {
	h.logger.Info("Handling end chat command", zap.Int64("chatID", cmd.ChatID), zap.Int64("userID", cmd.UserID))

	session, err := h.sessionRepo.GetSession(cmd.ChatID, cmd.UserID)
	if err != nil {
		return "", err
	}

	if !session.IsActive {
		return "ℹ️ Сессия общения уже не активна.", nil
	}

	session.IsActive = false
	session.Reset()

	if err := h.sessionRepo.SaveSession(session); err != nil {
		return "", err
	}

	return "👋 Сессия общения завершена. Контекст очищен.", nil
}

func (h *CommandHandler) HandleWhoAmI(ctx context.Context, cmd commands.WhoAmICommand) (string, error) {
	h.logger.Info("Handling whoami command", zap.Int64("chatID", cmd.ChatID), zap.Int64("userID", cmd.UserID))

	isActive := h.sessionRepo.IsSessionActive(cmd.ChatID, cmd.UserID)
	status := "неактивна"
	if isActive {
		status = "активна"
	}

	name := cmd.FirstName
	if cmd.LastName != "" {
		name += " " + cmd.LastName
	}

	return fmt.Sprintf(
		"👤 Информация о пользователе:\n"+
			"Имя: %s\n"+
			"Username: @%s\n"+
			"User ID: %d\n"+
			"Chat ID: %d\n"+
			"Сессия: %s",
		name, cmd.Username, cmd.UserID, cmd.ChatID, status,
	), nil
}

func (h *CommandHandler) HandleMessage(ctx context.Context, cmd commands.ProcessMessageCommand) (string, error) {
	h.logger.Info("Handling message", zap.Int64("chatID", cmd.ChatID), zap.Int64("userID", cmd.UserID))

	session, err := h.sessionRepo.GetSession(cmd.ChatID, cmd.UserID)
	if err != nil {
		return "", err
	}

	if !session.IsActive {
		return "ℹ️ Сессия не активна. Используй /startChat чтобы начать общение.", nil
	}

	// Добавляем сообщение пользователя
	session.AddMessage("user", cmd.Message)

	// Проверяем размер контекста
	if session.GetContextSize() > MaxContextSize {
		session.Reset()
		if err := h.sessionRepo.SaveSession(session); err != nil {
			return "", err
		}
		return "⚠️ Контекст стал слишком большим и был очищен. Пожалуйста, повтори свой вопрос.", nil
	}

	// Генерируем ответ
	response, err := h.claudeService.GenerateResponse(session.Messages)
	if err != nil {
		h.logger.Error("Failed to generate response", zap.Error(err))
		return "😔 Произошла ошибка при генерации ответа. Попробуй позже.", nil
	}

	// Добавляем ответ ассистента
	session.AddMessage("assistant", response)

	if err := h.sessionRepo.SaveSession(session); err != nil {
		return "", err
	}

	return response, nil
}

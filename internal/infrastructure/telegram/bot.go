package telegram

import (
	"context"
	"strings"
	"telegram-chatbot/internal/application/handlers"
	"telegram-chatbot/internal/config"
	"telegram-chatbot/internal/domain/commands"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type Bot struct {
	api            *tgbotapi.BotAPI
	config         *config.Config
	commandHandler *handlers.CommandHandler
	logger         *zap.Logger
}

func NewBot(
	config *config.Config,
	commandHandler *handlers.CommandHandler,
	logger *zap.Logger,
) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(config.TelegramBotToken)
	if err != nil {
		return nil, err
	}

	bot.Debug = false

	if err := setBotCommands(bot); err != nil {
		logger.Warn("Failed to set bot commands", zap.Error(err))
	}

	return &Bot{
		api:            bot,
		config:         config,
		commandHandler: commandHandler,
		logger:         logger,
	}, nil
}

func setBotCommands(bot *tgbotapi.BotAPI) error {
	commands := []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "Перезапустить бота и показать справку",
		},
		{
			Command:     "help",
			Description: "Показать справку по командам",
		},
		{
			Command:     "startChat",
			Description: "Начать сессию общения с контекстом",
		},
		{
			Command:     "endChat",
			Description: "Завершить сессию и очистить контекст",
		},
		{
			Command:     "whoami",
			Description: "Информация о пользователе и группе",
		},
	}

	config := tgbotapi.NewSetMyCommands(commands...)
	_, err := bot.Request(config)
	return err
}

func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("Starting bot", zap.String("username", b.api.Self.UserName))

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("Bot stopping...")
			return nil
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			go b.handleUpdate(ctx, update)
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	message := update.Message

	// Проверяем, что сообщение из разрешенной группы
	if message.Chat.ID != b.config.AllowedChatID {
		b.logger.Warn("Message from unauthorized chat", zap.Int64("chatID", message.Chat.ID))
		return
	}

	userID := message.From.ID
	chatID := message.Chat.ID

	var response string
	var err error

	// Обработка команд
	if message.IsCommand() {
		switch message.Command() {
		case "start":
			response, err = b.commandHandler.HandleStart(ctx, commands.StartCommand{
				ChatID: chatID,
				UserID: userID,
			})
		case "help":
			response, err = b.commandHandler.HandleHelp(ctx, commands.HelpCommand{
				ChatID: chatID,
				UserID: userID,
			})
		case "startChat":
			response, err = b.commandHandler.HandleStartChat(ctx, commands.StartChatCommand{
				ChatID: chatID,
				UserID: userID,
			})
		case "endChat":
			response, err = b.commandHandler.HandleEndChat(ctx, commands.EndChatCommand{
				ChatID: chatID,
				UserID: userID,
			})
		case "whoami":
			response, err = b.commandHandler.HandleWhoAmI(ctx, commands.WhoAmICommand{
				ChatID:    chatID,
				UserID:    userID,
				Username:  message.From.UserName,
				FirstName: message.From.FirstName,
				LastName:  message.From.LastName,
			})
		default:
			return // Неизвестная команда - игнорируем
		}
	} else {
		// Обработка обычных сообщений (только если бот упомянут)
		if b.isFromGroup(message) && !b.isBotMentioned(message) {
			return
		}

		b.sendTypingAction(chatID)

		response, err = b.commandHandler.HandleMessage(ctx, commands.ProcessMessageCommand{
			ChatID:   chatID,
			UserID:   userID,
			Message:  b.cleanMessage(message.Text),
			Username: message.From.UserName,
		})
	}

	if err != nil {
		b.logger.Error("Failed to handle message", zap.Error(err))
		response = "😔 Произошла ошибка. Попробуй позже."
	}

	if response != "" {
		msg := tgbotapi.NewMessage(chatID, response)
		msg.ReplyToMessageID = message.MessageID
		msg.DisableNotification = true

		if _, err := b.api.Send(msg); err != nil {
			b.logger.Error("Failed to send message", zap.Error(err))
		}
	}
}

func (b *Bot) sendTypingAction(chatID int64) {
	action := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	if _, err := b.api.Send(action); err != nil {
		b.logger.Debug("Failed to send typing action", zap.Error(err))
	}
}

func (b *Bot) isFromGroup(message *tgbotapi.Message) bool {
	return message.Chat.IsGroup() || message.Chat.IsSuperGroup()
}

func (b *Bot) isBotMentioned(message *tgbotapi.Message) bool {
	botUsername := "@" + b.api.Self.UserName
	return strings.Contains(message.Text, botUsername)
}

func (b *Bot) cleanMessage(text string) string {
	botUsername := "@" + b.api.Self.UserName
	return strings.TrimSpace(strings.ReplaceAll(text, botUsername, ""))
}

//go:build wireinject
// +build wireinject

package di

import (
	"telegram-chatbot/internal/application/handlers"
	"telegram-chatbot/internal/config"
	"telegram-chatbot/internal/domain/repositories"
	"telegram-chatbot/internal/domain/services"
	infraRepo "telegram-chatbot/internal/infrastructure/repositories"
	infraServices "telegram-chatbot/internal/infrastructure/services"
	"telegram-chatbot/internal/infrastructure/telegram"

	"github.com/google/wire"
	"go.uber.org/zap"
)

type Container struct {
	Bot *telegram.Bot
}

func InitializeContainer(*config.Config) (*Container, func(), error) {
	wire.Build(
		NewLogger,
		NewRedisSessionRepository,
		NewClaudeAPIService,
		handlers.NewCommandHandler,
		telegram.NewBot,
		wire.Struct(new(Container), "*"),
	)
	return &Container{}, nil, nil
}

func NewRedisSessionRepository(cfg *config.Config) repositories.SessionRepository {
	return infraRepo.NewRedisSessionRepository(cfg)
}

func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	config := zap.NewProductionConfig()

	switch cfg.LogLevel {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	}

	return config.Build()
}

func NewClaudeAPIService(cfg *config.Config) services.ClaudeService {
	return infraServices.NewClaudeAPIService(cfg.ClaudeAPIKey)
}

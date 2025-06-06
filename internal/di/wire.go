//go:build wireinject
// +build wireinject

package di

import (
	"telegram-chatbot/internal/application/handlers"
	"telegram-chatbot/internal/config"
	"telegram-chatbot/internal/domain/repositories"
	"telegram-chatbot/internal/domain/services"
	infraHealth "telegram-chatbot/internal/infrastructure/healthcheck"
	infraRepo "telegram-chatbot/internal/infrastructure/repositories"
	infraServices "telegram-chatbot/internal/infrastructure/services"
	"telegram-chatbot/internal/infrastructure/telegram"
	"telegram-chatbot/internal/presentation/controllers"

	"github.com/google/wire"
	"go.uber.org/zap"
)

type Container struct {
	Bot         *telegram.Bot
	HealthCheck *infraHealth.Service
}

func InitializeContainer(*config.Config) (*Container, func(), error) {
	wire.Build(
		NewLogger,
		NewRedisSessionRepository,
		NewClaudeAPIService,
		handlers.NewCommandHandler,
		telegram.NewBot,
		NewHealthCheckService,
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

func NewHealthCheckService(cfg *config.Config, logger *zap.Logger) *infraHealth.Service {
	srv := infraHealth.NewHealthCheckService(logger, cfg.HealthCheckPort)
	healthCtrl := controllers.NewHealthCheckController(srv.ReadyFlag())
	docsCtrl := controllers.NewDocumentationController()
	healthCtrl.RegisterRoutes(srv.Router().Group("/health"))
	docsCtrl.RegisterRoutes(srv.Router().Group("/swagger"))
	return srv
}

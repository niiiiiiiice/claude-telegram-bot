package app

import (
	"context"
	"telegram-chatbot/internal/di"
)

type App struct {
	container *di.Container
}

func New(container *di.Container) *App {
	return &App{
		container: container,
	}
}

func (a *App) Run(ctx context.Context) error {
	return a.container.Bot.Start(ctx)
}

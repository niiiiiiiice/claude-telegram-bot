package app

import (
	"context"
	"sync"
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
	// Create a wait group to wait for both services
	var wg sync.WaitGroup
	wg.Add(2)

	// Create an error channel to collect errors
	errCh := make(chan error, 2)

	// Start the bot in a goroutine
	go func() {
		defer wg.Done()
		if err := a.container.Bot.Start(ctx); err != nil {
			errCh <- err
		}
	}()

	// Start the health check service in a goroutine
	go func() {
		defer wg.Done()
		if err := a.container.HealthCheck.Start(ctx); err != nil {
			errCh <- err
		}
	}()

	// Wait for both services to complete or for an error
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Return the first error if any
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

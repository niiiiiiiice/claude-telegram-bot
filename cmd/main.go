package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	docs "telegram-chatbot/docs"
	"telegram-chatbot/internal/app"
	"telegram-chatbot/internal/config"
	"telegram-chatbot/internal/di"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем .env файл (игнорируем ошибки для продакшена)
	_ = godotenv.Load(".env.example")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	docs.SwaggerInfo.BasePath = "/"

	container, cleanup, err := di.InitializeContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}
	defer cleanup()

	application := app.New(container)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
	}()

	if err := application.Run(ctx); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

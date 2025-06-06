package healthcheck

import (
	"context"
	"net/http"
	"sync/atomic"
	"telegram-chatbot/internal/infrastructure/telegram"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// Service represents the health check service
type Service struct {
	router *gin.Engine
	bot    *telegram.Bot
	logger *zap.Logger
	port   string
	ready  atomic.Bool
}

// NewHealthCheckService creates a new health check service
func NewHealthCheckService(bot *telegram.Bot, logger *zap.Logger, port string) *Service {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	service := &Service{
		router: router,
		bot:    bot,
		logger: logger,
		port:   port,
	}

	// Set ready to false initially
	service.ready.Store(false)

	// Register routes
	router.GET("/health/liveness", service.livenessHandler)
	router.GET("/health/readiness", service.readinessHandler)
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return service
}

// Start starts the health check service
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Starting health check service", zap.String("port", s.port))

	// Mark as ready after the bot has started
	s.ready.Store(true)

	server := &http.Server{
		Addr:    ":" + s.port,
		Handler: s.router,
	}

	// Run server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Failed to start health check server", zap.Error(err))
		}
	}()

	// Wait for context cancellation to shut down
	<-ctx.Done()
	s.logger.Info("Shutting down health check service")

	// Create a timeout context for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Health check server shutdown error", zap.Error(err))
		return err
	}

	return nil
}

// SetReady marks the service as ready
func (s *Service) SetReady(ready bool) {
	s.ready.Store(ready)
}

// livenessHandler handles liveness probe requests
// @Summary Liveness check
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health/liveness [get]
func (s *Service) livenessHandler(c *gin.Context) {
	// Liveness probe just checks if the service is running
	c.JSON(http.StatusOK, gin.H{
		"status": "UP",
	})
}

// readinessHandler handles readiness probe requests
// @Summary Readiness check
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /health/readiness [get]
func (s *Service) readinessHandler(c *gin.Context) {
	// Readiness probe checks if the bot is ready to handle requests
	if s.ready.Load() {
		c.JSON(http.StatusOK, gin.H{
			"status": "READY",
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "NOT_READY",
		})
	}
}

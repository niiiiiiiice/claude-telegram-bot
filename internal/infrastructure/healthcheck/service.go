package healthcheck

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Service represents the health check service
type Service struct {
	router *gin.Engine
	logger *zap.Logger
	port   string
	ready  atomic.Bool
}

// NewHealthCheckService creates a new health check service
func NewHealthCheckService(logger *zap.Logger, port string) *Service {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	service := &Service{
		router: router,
		logger: logger,
		port:   port,
	}

	// Set ready to false initially
	service.ready.Store(false)

	return service
}

// Router returns the underlying HTTP router for registering routes.
func (s *Service) Router() *gin.Engine {
	return s.router
}

// ReadyFlag exposes the readiness indicator used by controllers.
func (s *Service) ReadyFlag() *atomic.Bool {
	return &s.ready
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
		fmt.Println("Starting server at http://localhost:" + s.port)
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

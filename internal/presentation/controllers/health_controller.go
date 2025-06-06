package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"sync/atomic"
)

// HealthCheckController handles health related endpoints.
type HealthCheckController struct {
	ready *atomic.Bool
}

// NewHealthCheckController creates a new HealthCheckController.
func NewHealthCheckController(ready *atomic.Bool) *HealthCheckController {
	return &HealthCheckController{ready: ready}
}

// RegisterRoutes registers health check routes under the provided router group.
func (h *HealthCheckController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/liveness", h.liveness)
	rg.GET("/readiness", h.readiness)
}

// liveness godoc
// @Summary Liveness check
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health/liveness [get]
func (h *HealthCheckController) liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "UP"})
}

// readiness godoc
// @Summary Readiness check
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /health/readiness [get]
func (h *HealthCheckController) readiness(c *gin.Context) {
	if h.ready.Load() {
		c.JSON(http.StatusOK, gin.H{"status": "READY"})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "NOT_READY"})
	}
}

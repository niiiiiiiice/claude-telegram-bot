package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// DocumentationController serves swagger documentation.
type DocumentationController struct{}

// NewDocumentationController creates a new DocumentationController.
func NewDocumentationController() *DocumentationController {
	return &DocumentationController{}
}

// RegisterRoutes registers documentation routes under the provided router group.
func (d *DocumentationController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, c.FullPath()+"/index.html")
	})
	rg.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

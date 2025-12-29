package http

import (
	"github.com/Ebrahim-hamdy/mastara-saas/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes sets up the routes for the IAM module.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// These routes are grouped under a path like /api/v1
	// The Authenticator middleware is applied to this group by the main router.
	authGroup := router.Group("/auth")
	{
		// Note: Login and Register are conceptually "auth" actions, but registration
		// is performed by an already authenticated admin user in our current flow.
		authGroup.POST("/register", middleware.ErrorHandler(h.RegisterUser))
		authGroup.POST("/login", middleware.ErrorHandler(h.LoginUser))
	}
}

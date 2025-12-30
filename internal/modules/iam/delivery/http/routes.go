package http

import (
	"github.com/Ebrahim-hamdy/mastara-saas/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes sets up the public-facing routes for the IAM module (e.g., login).
func (h *Handler) RegisterPublicRoutes(router *gin.RouterGroup) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", middleware.ErrorHandler(h.LoginEmployee))
		// The "Accept Invite" and "Set Password" routes would also be public.
		// authGroup.POST("/accept-invite", middleware.ErrorHandler(h.AcceptInvite))
	}
}

// RegisterProtectedRoutes sets up the protected, staff-only routes for the IAM module.
func (h *Handler) RegisterProtectedRoutes(router *gin.RouterGroup) {
	// All routes in this group are protected by the Authenticator middleware.
	employeesGroup := router.Group("/employees")
	{
		// POST /api/v1/employees/invite - Invite a new staff member.
		employeesGroup.POST("/invite", middleware.ErrorHandler(h.InviteEmployee))
		// Other employee management routes (GET /, GET /:id, PUT /:id) would go here.
	}
}

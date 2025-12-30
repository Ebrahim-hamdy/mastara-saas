package http

import (
	"github.com/Ebrahim-hamdy/mastara-saas/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes sets up the routes for the Patient module.
// All these routes are protected and require an authenticated staff member.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	patientGroup := router.Group("/patients")
	{
		// POST /api/v1/patients - Create a new, fully registered patient
		patientGroup.POST("/", middleware.ErrorHandler(h.RegisterPatient))

		// PUT /api/v1/patients/:id/complete-registration - Upgrade a guest to registered
		patientGroup.PUT("/:id/complete-registration", middleware.ErrorHandler(h.CompleteGuestProfile))

		patientGroup.GET("/", middleware.ErrorHandler(h.ListPatients))
		patientGroup.GET("/:id", middleware.ErrorHandler(h.GetPatient))

		// We can add a DELETE "/:id" for archiving later.
	}

	// === PUBLIC ROUTES (NO AUTH) ===
	// public := router.Group("/public")
	{
		// Here we would register the public booking and guest management handlers
		// public.POST("/appointments", middleware.ErrorHandler(appointmentHandler.CreateGuestAppointment))
		// public.GET("/appointments/manage", middleware.GuestTokenAuthenticator(), middleware.ErrorHandler(appointmentHandler.GetGuestAppointment))
		// public.POST("/appointments/manage/cancel", middleware.GuestTokenAuthenticator(), middleware.ErrorHandler(appointmentHandler.CancelGuestAppointment))
	}
}

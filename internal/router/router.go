// Package router is responsible for defining and configuring all the HTTP routes for the application.
package router

import (
	"context"
	"net/http"
	"time"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/infra/database"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/infra/security"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/middleware" // <-- Import new middleware
	iamHttp "github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam/delivery/http"
	"github.com/Ebrahim-hamdy/mastara-saas/pkg/apierror" // <-- Import new apierror
	"github.com/gin-gonic/gin"
)

// New creates and returns a new Gin engine with all the application routes configured.
func New(dbProvider *database.Provider, tokenManager *security.PasetoManager, iamHandler *iamHttp.Handler) *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.BodyLimiter(1_048_576)) // 1MB limit

	// Health check handler now uses our centralized error handler.
	router.GET("/health", middleware.ErrorHandler(healthCheckHandler(dbProvider)))

	v1 := router.Group("/api/v1")
	v1.Use(middleware.Authenticator(tokenManager))
	{
		iamHandler.RegisterRoutes(v1)

		// Example of a protected route
		v1.GET("/me", func(c *gin.Context) {
			payload := middleware.GetAuthPayload(c.Request.Context())
			if payload == nil {
				// This should technically be unreachable if middleware is working
				err := apierror.NewUnauthorized("Not authenticated", nil)
				c.AbortWithStatusJSON(err.StatusCode, gin.H{"error": err.PublicMessage})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": payload})
		})

		// In Stage 1, we will register module routes here:
		// iam.RegisterRoutes(v1, dbProvider.Pool)
	}

	return router
}

// healthCheckHandler now returns an *apierror.APIError, simplifying its logic.
func healthCheckHandler(db *database.Provider) middleware.APIHandlerFunc {
	return func(c *gin.Context) *apierror.APIError {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := db.HealthCheck(ctx); err != nil {
			// Instead of writing a JSON response here, we just return our structured error.
			// The middleware will handle the rest.
			return apierror.NewInternalServer(err)
		}

		c.JSON(200, gin.H{"status": "healthy"})
		return nil // On success, return nil.
	}
}

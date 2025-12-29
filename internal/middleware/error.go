// Package middleware provides HTTP middleware functions for cross-cutting concerns.
package middleware

import (
	"github.com/Ebrahim-hamdy/mastara-saas/pkg/apierror"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// APIHandlerFunc is a custom handler function that can return an APIError.
type APIHandlerFunc func(c *gin.Context) *apierror.APIError

// ErrorHandler is a middleware that adapts an APIHandlerFunc to a standard gin.HandlerFunc.
// It centrally handles the logic for logging internal errors and sending a clean JSON response to the client.
func ErrorHandler(h APIHandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := h(c); err != nil {
			// Log the internal, detailed error for debugging.
			// The public message is intentionally not logged here as it's for the client.
			log.Error().
				Err(err). // This logs the full internal error chain
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Int("status_code", err.StatusCode).
				Msg("API error occurred")

			// Send a structured, public-facing error response to the client.
			c.AbortWithStatusJSON(err.StatusCode, gin.H{
				"error": gin.H{
					"message": err.PublicMessage,
					"code":    err.StatusCode,
				},
			})
		}
	}
}

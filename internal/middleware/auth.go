package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/infra/security"
	"github.com/Ebrahim-hamdy/mastara-saas/pkg/apierror"
	"github.com/gin-gonic/gin"
)

// contextKey is an unexported type to be used as a key for context values.
// This prevents collisions with other packages.
type contextKey string

const (
	authPayloadKey = contextKey("auth_payload")
	// Error message constant
	ErrAuthPayloadNotFoundMsg = "auth payload not found in context"
)

// Authenticator is a middleware that verifies the authentication token and injects
// the security context (AuthPayload) into the request.
func Authenticator(tokenManager *security.PasetoManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			err := apierror.NewUnauthorized("authorization header is required", nil)
			c.AbortWithStatusJSON(err.StatusCode, gin.H{"error": err.PublicMessage})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			err := apierror.NewUnauthorized("invalid authorization header format", nil)
			c.AbortWithStatusJSON(err.StatusCode, gin.H{"error": err.PublicMessage})
			return
		}

		token := parts[1]
		payload, err := tokenManager.VerifyToken(token)
		if err != nil {
			apiErr := apierror.NewUnauthorized("invalid or expired token", err)
			c.AbortWithStatusJSON(apiErr.StatusCode, gin.H{"error": apiErr.PublicMessage})
			return
		}

		// Inject the payload into the request context.
		ctx := context.WithValue(c.Request.Context(), authPayloadKey, payload)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetAuthPayload retrieves the authenticated user's payload from the context.
// It returns nil if the payload is not present.
func GetAuthPayload(ctx context.Context) (*security.AuthPayload, error) {
	payload, ok := ctx.Value(authPayloadKey).(*security.AuthPayload)
	if !ok || payload == nil {
		return nil, errors.New(ErrAuthPayloadNotFoundMsg)
	}
	return payload, nil
}

package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const oneYearInSeconds = 31536000

// SecurityHeaders applies a set of security-related HTTP headers to every response.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevents clickjacking
		c.Header("X-Frame-Options", "DENY")
		// Prevents browsers from MIME-sniffing a response away from the declared content-type
		c.Header("X-Content-Type-Options", "nosniff")
		// Enables the XSS filter built into most recent web browsers
		c.Header("X-XSS-Protection", "1; mode=block")
		// Enforces HTTPS
		c.Header("Strict-Transport-Security", fmt.Sprintf("max-age=%d; includeSubDomains", oneYearInSeconds))
		// A basic Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none';")

		c.Next()
	}
}

// BodyLimiter restricts the size of incoming request bodies to prevent DoS attacks.
func BodyLimiter(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}

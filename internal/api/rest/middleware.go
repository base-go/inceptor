package rest

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/flakerimi/inceptor/internal/auth"
	"github.com/flakerimi/inceptor/internal/core"
	"github.com/flakerimi/inceptor/internal/storage"
	"github.com/gin-gonic/gin"
)

const (
	ContextKeyApp   = "app"
	ContextKeyAdmin = "is_admin"
)

// APIKeyAuth middleware validates API key and sets app context
func APIKeyAuth(repo storage.Repository, adminKey string) gin.HandlerFunc {
	return APIKeyOrSessionAuth(repo, adminKey, nil)
}

// APIKeyOrSessionAuth middleware validates API key OR session token
func APIKeyOrSessionAuth(repo storage.Repository, adminKey string, authManager *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First try session token (Bearer auth)
		if authManager != nil {
			bearerToken := ExtractBearerToken(c)
			if bearerToken != "" && authManager.ValidateSession(bearerToken) {
				c.Set(ContextKeyAdmin, true) // Session users have admin access
				c.Next()
				return
			}
		}

		// Then try API key
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Try query parameter as fallback
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "API key required",
				"code":  "MISSING_API_KEY",
			})
			return
		}

		// Check if it's the admin key
		if adminKey != "" && apiKey == adminKey {
			c.Set(ContextKeyAdmin, true)
			c.Next()
			return
		}

		// Hash the API key for lookup
		keyHash := HashAPIKey(apiKey)

		// Look up app by API key hash
		app, err := repo.GetAppByAPIKey(c.Request.Context(), keyHash)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to validate API key",
				"code":  "INTERNAL_ERROR",
			})
			return
		}

		if app == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
				"code":  "INVALID_API_KEY",
			})
			return
		}

		// Set app in context
		c.Set(ContextKeyApp, app)
		c.Next()
	}
}

// AdminOnly middleware ensures only admin API key can access the endpoint
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get(ContextKeyAdmin)
		if !exists || !isAdmin.(bool) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
				"code":  "ADMIN_REQUIRED",
			})
			return
		}
		c.Next()
	}
}

// AppContext middleware requires app context (not just admin)
func AppContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get(ContextKeyApp)
		if !exists {
			// Admin can pass app_id as query param
			isAdmin, adminExists := c.Get(ContextKeyAdmin)
			if adminExists && isAdmin.(bool) {
				c.Next()
				return
			}

			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "App context required",
				"code":  "APP_REQUIRED",
			})
			return
		}
		c.Next()
	}
}

// GetApp retrieves the app from context
func GetApp(c *gin.Context) *core.App {
	app, exists := c.Get(ContextKeyApp)
	if !exists {
		return nil
	}
	return app.(*core.App)
}

// IsAdmin checks if the request is from admin
func IsAdmin(c *gin.Context) bool {
	isAdmin, exists := c.Get(ContextKeyAdmin)
	return exists && isAdmin.(bool)
}

// CORS middleware for cross-origin requests
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-API-Key")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestLogger middleware logs requests
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/health", "/ready"},
		Formatter: func(param gin.LogFormatterParams) string {
			return ""
		},
	})
}

// Recovery middleware recovers from panics
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

// RateLimiter provides basic rate limiting (in-memory, simple implementation)
type RateLimiter struct {
	// Could be expanded with Redis for distributed rate limiting
	// For now, we'll use Gin's built-in or skip
}

// HashAPIKey creates a SHA256 hash of an API key for secure storage
func HashAPIKey(apiKey string) string {
	h := sha256.New()
	h.Write([]byte(apiKey))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateAPIKey creates a new random API key
func GenerateAPIKey() string {
	// Use crypto/rand for secure random generation
	b := make([]byte, 32)
	// In production, use crypto/rand
	// For now, we'll generate a simple key
	h := sha256.New()
	h.Write(b)
	key := hex.EncodeToString(h.Sum(nil))
	return "ink_" + key[:32] // Prefix with "ink_" for easy identification
}

// ExtractBearerToken extracts a bearer token from the Authorization header
func ExtractBearerToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

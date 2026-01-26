package rest

import (
	"net/http"

	"github.com/flakerimi/inceptor/internal/auth"
	"github.com/gin-gonic/gin"
)

// AuthHandler holds auth-related handlers
type AuthHandler struct {
	authManager *auth.Manager
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authManager *auth.Manager) *AuthHandler {
	return &AuthHandler{authManager: authManager}
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=4"`
}

// Status returns auth status
func (h *AuthHandler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"needs_password_change": h.authManager.NeedsPasswordChange(),
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	if !h.authManager.ValidatePassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	session, err := h.authManager.CreateSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":                 session.Token,
		"expires_at":            session.ExpiresAt,
		"needs_password_change": h.authManager.NeedsPasswordChange(),
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token != "" {
		// Remove "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
		h.authManager.DeleteSession(token)
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if !h.authManager.ChangePassword(req.OldPassword, req.NewPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid old password or new password too short"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// SessionAuth middleware validates session token
func SessionAuth(authManager *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No authorization token"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		if !authManager.ValidateSession(token) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
			c.Abort()
			return
		}

		c.Next()
	}
}

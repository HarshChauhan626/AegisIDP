package middleware

import (
	"net/http"
	"strings"

	"github.com/HarshChauhan626/AegisIDP/backend/auth"
	"github.com/gin-gonic/gin"
)

const (
	ctxKeyUserID = "user_id"
	ctxKeyRole   = "role"
)

// JWTAuth validates the Bearer token on protected routes and injects user claims into the context.
func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Authorization header is required",
			}})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Authorization header must be in the format: Bearer <token>",
			}})
			return
		}

		claims, err := auth.ValidateToken(parts[1], jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{
				"code":    "TOKEN_INVALID",
				"message": "Invalid or expired token",
			}})
			return
		}

		c.Set(ctxKeyUserID, claims.UserID)
		c.Set(ctxKeyRole, claims.Role)
		c.Next()
	}
}

// GetUserID extracts the authenticated user ID from the Gin context.
func GetUserID(c *gin.Context) string {
	id, _ := c.Get(ctxKeyUserID)
	if s, ok := id.(string); ok {
		return s
	}
	return ""
}

// GetRole extracts the authenticated user role from the Gin context.
func GetRole(c *gin.Context) string {
	role, _ := c.Get(ctxKeyRole)
	if s, ok := role.(string); ok {
		return s
	}
	return ""
}

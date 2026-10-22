package auth

import (
	"net/http"
	"strings"

	"backend/internal/redis"

	"github.com/gin-gonic/gin"
)

const (
	ContextClaimsKey = "auth_claims"
	ContextTokenKey  = "auth_token"
)

// AuthRequired ensures that the request carries a valid, non-blacklisted Bearer token.
func AuthRequired(secret string, store *redis.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format; expected 'Bearer <token>'"})
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "bearer token is missing"})
			return
		}

		if store != nil && store.IsTokenBlacklisted(c.Request.Context(), tokenString) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
			return
		}

		claims, err := ValidateToken(tokenString, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ContextClaimsKey, claims)
		c.Set(ContextTokenKey, tokenString)
		c.Next()
	}
}

// RequireRoles restricts access to users with one of the specified roles.
func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := GetClaims(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}

		for _, role := range allowedRoles {
			if claims.Role == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "permission denied: insufficient permissions for this action",
		})
	}
}

// GetClaims retrieves the authenticated user's Claims from the Gin context.
func GetClaims(c *gin.Context) (*Claims, bool) {
	val, exists := c.Get(ContextClaimsKey)
	if !exists {
		return nil, false
	}
	claims, ok := val.(*Claims)
	return claims, ok
}

// GetToken retrieves the raw bearer token string from the Gin context.
func GetToken(c *gin.Context) (string, bool) {
	val, exists := c.Get(ContextTokenKey)
	if !exists {
		return "", false
	}
	token, ok := val.(string)
	return token, ok
}

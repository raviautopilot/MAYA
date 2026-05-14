// Package middleware provides Gin middleware for JWT authentication and error recovery.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"mykanban-backend/auth"
	"mykanban-backend/models"
)

// JWTAuth returns a Gin middleware that validates JWT tokens from the Authorization header.
func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
				Error:  "missing authorization header",
				Status: http.StatusUnauthorized,
			})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
				Error:  "invalid authorization header format, expected 'Bearer <token>'",
				Status: http.StatusUnauthorized,
			})
			return
		}

		claims, err := auth.ValidateJWT(parts[1], jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
				Error:  "invalid or expired token: " + err.Error(),
				Status: http.StatusUnauthorized,
			})
			return
		}

		c.Set("email", claims.Email)
		c.Next()
	}
}

// Recovery returns a Gin middleware that recovers from panics and returns a JSON error.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.APIResponse{
			Error:  "internal server error",
			Status: http.StatusInternalServerError,
		})
	})
}

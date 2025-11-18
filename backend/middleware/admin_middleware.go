package middleware

import (
	"net/http"

	"Personal-Finance-Tracker-backend/db"
	"Personal-Finance-Tracker-backend/models"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// RequireAdmin middleware checks if the user has admin role
func RequireAdmin() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// First check if user is authenticated
		claims, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		jwtClaims := claims.(jwt.MapClaims)
		userID := uint(jwtClaims["sub"].(float64))

		// Check role from JWT token first (if available)
		if role, ok := jwtClaims["role"]; ok {
			if role.(string) == string(models.UserRoleAdmin) {
				// Get user from database for context
				var user models.User
				if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
					c.Abort()
					return
				}
				c.Set("adminUser", user)
				c.Next()
				return
			} else {
				c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
				c.Abort()
				return
			}
		}

		// Fallback: Get user from database to check role (for older tokens without role)
		var user models.User
		if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			c.Abort()
			return
		}

		// Check if user has admin role
		if user.Role != models.UserRoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}

		// Store user object in context for later use
		c.Set("adminUser", user)
		c.Next()
	})
}

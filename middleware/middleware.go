package middleware

import (
	"go-be/database"
	"go-be/models"
	"go-be/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware untuk validasi JWT
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			c.Abort()
			return
		}

		userID, err := utils.ReverseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error parse token"})
			c.Abort()
			return
		}
		c.Set("userId", userID)
		println(userID)
		c.Next()
	}
}

func AdminMiddleware(c *gin.Context) {
	var user models.User
	userID, exists := c.Get("userId")
	if !exists {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Admin access only"})
		return
	}
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}
	if user.Role != "admin" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Admin access only"})
		return
	}
	c.Next()
}

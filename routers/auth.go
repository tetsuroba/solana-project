package routers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"solana/utils"
)

// AuthMiddleware validates the JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		claims, valid := utils.ValidateToken(tokenString)
		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("username", claims.Username)

		c.Next()
	}
}

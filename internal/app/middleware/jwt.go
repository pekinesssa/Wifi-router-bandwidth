package middleware

import (
	"net/http"
	"strings"

	"Wi-Fi-router-bandwidth-backend/internal/pkg"

	"github.com/gin-gonic/gin"
)

func AuthRequired(redisClient *pkg.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизирвоаться"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неправильный токен"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		if redisClient != nil {
			inBlacklist, err := redisClient.IsInBlacklist(c.Request.Context(), tokenString)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
				c.Abort()
				return
			}
			if inBlacklist {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "ТОкен был добавлен в блэклист"})
				c.Abort()
				return
			}
		}

		claims, err := pkg.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("is_moderator", claims.IsModerator)
		c.Set("login", claims.Login)
		c.Next()
	}
}


func ModeratorRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		isModerator, exists := c.Get("is_moderator")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "---"})
			c.Abort()
			return
		}

		if !isModerator.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Требуются права модератора"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func GetUserID(c *gin.Context) uint {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return userID.(uint)
}
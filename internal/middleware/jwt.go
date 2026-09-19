package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jekyulll/url_shortener/pkg/jwt"
)

func JWTAuther(jwt *jwt.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取Authorization字段
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}
		// 分割Authorization字段
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			return
		}

		tokenString := parts[1]
		// 解析JWT token
		claims, err := jwt.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}

		c.Set("email", claims.Email)
		c.Set("userID", claims.UserID)

		c.Next()
	}
}

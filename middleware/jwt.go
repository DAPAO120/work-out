package middleware

import (
	"Project001/common/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {

	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg": "missing token",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")

		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg": "invalid token format",
			})
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(parts[1])

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg": "token invalid",
			})
			c.Abort()
			return
		}

		// 保存用户ID到上下文
		c.Set("user_id", claims.UserID)
		c.Set("is_admin", claims.IsAdmin) // 这里设置 is_admin

		c.Next()
	}
}

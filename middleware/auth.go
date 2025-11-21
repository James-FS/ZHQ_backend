package middleware

import (
	"strings"
	"zhq-backend/utils"

	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "Authorization header required")
			c.Abort() //中止当前请求，不再执行后续的中间件和处理函数
			return
		}

		// 检查Bearer token格式
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			utils.Unauthorized(c, "Invalid authorization format")
			c.Abort()
			return
		}

		// 验证JWT逻辑
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			utils.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("open_id", claims.OpenID)
		c.Next()
	}
}

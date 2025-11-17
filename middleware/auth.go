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
			c.Abort()
			return
		}

		// 检查Bearer token格式
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			utils.Unauthorized(c, "Invalid authorization format")
			c.Abort()
			return
		}

		// 这里后续会添加JWT验证逻辑
		// 现在先简单验证token不为空
		if tokenString == "" {
			utils.Unauthorized(c, "Token is required")
			c.Abort()
			return
		}

		// 设置用户信息到上下文（后续实现JWT验证后会设置真实用户信息）
		c.Set("user_id", 1) // 临时设置，后续会从JWT中解析

		c.Next()
	}
}

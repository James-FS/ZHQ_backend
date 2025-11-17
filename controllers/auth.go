package controllers

import (
	"zhq-backend/utils"

	"github.com/gin-gonic/gin"
)

// 微信登录
func WeChatLogin(c *gin.Context) {
	// 获取请求参数
	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 这里后续会实现微信登录逻辑
	// 现在先返回模拟数据
	utils.Success(c, gin.H{
		"token": "mock_token_" + req.Code,
		"user": gin.H{
			"id":       1,
			"nickname": "测试用户",
			"avatar":   "https://example.com/avatar.jpg",
		},
	})
}

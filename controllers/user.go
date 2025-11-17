package controllers

import (
	"zhq-backend/utils"

	"github.com/gin-gonic/gin"
)

// 获取用户信息
func GetUserProfile(c *gin.Context) {
	// 从中间件中获取用户ID
	userID := c.GetInt("user_id")

	// 这里后续会从数据库查询用户信息
	// 现在先返回模拟数据
	utils.Success(c, gin.H{
		"id":       userID,
		"nickname": "测试用户",
		"avatar":   "https://example.com/avatar.jpg",
		"gender":   1,
		"phone":    "13800138000",
	})
}

// 更新用户信息
func UpdateUserProfile(c *gin.Context) {
	var req struct {
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
		Gender   int    `json:"gender"`
		Phone    string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 这里后续会实现更新用户信息的逻辑
	utils.SuccessWithMessage(c, "更新成功", nil)
}

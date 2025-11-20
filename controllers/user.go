package controllers

import (
	"zhq-backend/database"
	"zhq-backend/models"
	"zhq-backend/utils"

	"github.com/gin-gonic/gin"
)

func MockUser() models.User {
	return models.User{
		UserID:   "0",
		OpenID:   "test_openid_456",
		UnionID:  "test_unionid_789",
		Nickname: "测试用户",
		Avatar:   "https://example.com/avatar.jpg",
		Gender:   1,
		Phone:    "13800138000",
		Status:   1,
	}
}

// 获取用户信息
func GetUserProfile(c *gin.Context) {
	// 从中间件中获取用户ID
	userID := c.GetString("user_id")
	if userID == "" {
		utils.BadRequest(c, "用户未鉴权")
		return
	}

	var user models.User
	user = MockUser()
	// 这里后续会从数据库查询用户信息
	// 现在先返回模拟数据
	utils.Success(c, gin.H{
		"user": user,
	})
}

// 更新用户信息
func UpdateUserProfile(c *gin.Context) {

	var profileUpdate map[string]interface{}
	if err := c.ShouldBindJSON(&profileUpdate); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	var user models.User
	user = MockUser()

	userID := c.Param("user_id")
	//userID = "0"
	if userID == "" {
		utils.BadRequest(c, "userID不能为空")
		return
	}

	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		utils.BadRequest(c, "用户不存在")
		return
	}

	UpdateFields := map[string]bool{
		"nickname": true,
		"avatar":   true,
		"gender":   true,
		"phone":    true,
	}
	for key := range profileUpdate {
		if !UpdateFields[key] {
			delete(profileUpdate, key)
		}
	}

	if err := database.DB.Model(&user).Updates(profileUpdate).Error; err != nil {
		utils.BadRequest(c, "用户资料更新失败")
		return
	}

	utils.SuccessWithMessage(c, "更新用户资料成功", gin.H{
		"user": user,
	})
}

// 获取用户收藏
func GetUserCollection(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.BadRequest(c, "用户未鉴权")
		return
	}
	var collections []models.UserCollection
	var total int64
	page := 1
	pageSize := 10
	if err := database.DB.Where("user_id = ?", userID).
		Model(&models.UserCollection{}).
		Count(&total).Error; err != nil {
		utils.BadRequest(c, "查询收藏总数失败")
		return
	}

	if err := database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&collections).Error; err != nil {
		utils.BadRequest(c, "查询收藏列表失败")
		return
	}
	utils.Success(c, gin.H{
		"total": total,
		"list":  collections,
	})
}

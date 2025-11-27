package controllers

import (
	"errors"
	"zhq-backend/database"
	"zhq-backend/models"
	"zhq-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PhonePasswordLogin 手机号密码登录
func PhonePasswordLogin(c *gin.Context) {
	// 获取请求参数
	var req struct {
		Phone    string `json:"phone" binding:"required,len=11"` // 验证手机号长度
		Password string `json:"password" binding:"required,min=6,max=20"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 查询用户
	var user models.User
	db := database.GetDB()
	if err := db.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		utils.BadRequest(c, "手机号或密码错误")
		return
	}

	// 验证密码
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		utils.BadRequest(c, "手机号或密码错误")
		return
	}

	// 检查用户状态
	if user.Status != 1 {
		utils.BadRequest(c, "用户已被禁用")
		return
	}

	// 生成JWT令牌
	token, err := utils.GenerateToken(user.UserID, user.OpenID)
	if err != nil {
		utils.InternalServerError(c, "生成令牌失败: ", err)
		return
	}

	// 返回用户信息和令牌
	utils.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"user_id":  user.UserID,
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
			"gender":   user.Gender,
			"phone":    user.Phone,
		},
	})
}

// RegisterByPhone 手机号注册
func RegisterByPhone(c *gin.Context) {
	var req struct {
		Phone    string `json:"phone" binding:"required,len=11"`
		Password string `json:"password" binding:"required,min=6,max=20"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误:"+err.Error())
		return
	}

	//检查手机号是否已存在
	var existingUser models.User
	db := database.GetDB()
	err := db.Where("phone = ?", req.Phone).First(&existingUser).Error

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			// 数据库错误
			utils.InternalServerError(c, "查询用户失败", err)
			return
		}
		// 记录不存在，说明手机号未被注册，可以继续
	} else {
		// 找到了记录，说明手机号已被注册
		utils.BadRequest(c, "手机号已被注册")
		return
	}

	//密码加密
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.InternalServerError(c, "密码加密失败: ", err)
		return
	}

	//创建新用户
	user := models.User{
		UserID:   uuid.New().String(),
		Phone:    req.Phone,
		Password: hashedPassword,
		Status:   1,
		Tags:     "[]",
		OpenID:   "phone_" + uuid.New().String(),
	}

	if err := db.Create(&user).Error; err != nil {
		utils.InternalServerError(c, "创建用户失败:", err)
		return
	}

	utils.SuccessWithMessage(c, "注册成功", nil)
}

// WeChatLogin 微信登录
func WeChatLogin(c *gin.Context) {
	// 获取请求参数
	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	//调用微信接口获取openid
	wechatResp, err := utils.GetWeChatOpenID(req.Code)
	if err != nil {
		utils.BadRequest(c, "微信登录失败: "+err.Error())
		return
	}

	//检查用户是否已存在
	var user models.User
	db := database.GetDB()
	if err := db.Where("open_id = ?", wechatResp.OpenID).First(&user).Error; err != nil {

		//用户不存在，创建新用户
		user = models.User{
			UserID:  uuid.New().String(),
			OpenID:  wechatResp.OpenID,
			UnionID: wechatResp.UnionID,
			Status:  1,
		}
		if err := db.Create(&user).Error; err != nil {
			utils.InternalServerError(c, "创建用户失败: ", err)
			return
		}
	}

	//生成JWT令牌
	token, err := utils.GenerateToken(user.UserID, user.OpenID)
	if err != nil {
		utils.InternalServerError(c, "生成令牌失败: ", err)
		return
	}

	// 返回用户信息和令牌
	utils.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"user_id":  user.UserID,
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
			"gender":   user.Gender,
		},
	})
}

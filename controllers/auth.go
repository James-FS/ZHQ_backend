package controllers

import (
	"zhq-backend/database"
	"zhq-backend/models"
	"zhq-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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

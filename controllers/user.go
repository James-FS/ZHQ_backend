package controllers

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"zhq-backend/database"
	"zhq-backend/models"
	"zhq-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	var user models.User
	if userID == "" {
		utils.BadRequest(c, "用户未鉴权")
		return
	}

	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		utils.BadRequest(c, "用户不存在")
		return
	}

	//user = MockUser()
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

	userID := c.GetString("user_id")
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

// 上传用户头像
func UploadAvatar(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.BadRequest(c, "userID不能为空")
		return
	}
	avatar, err := c.FormFile("avatar")
	if err != nil {
		utils.BadRequest(c, "上传失败")
		return
	}
	if avatar.Size > 4*1024*1024 {
		utils.BadRequest(c, "上传图片应小于4MB")
	}
	ext := filepath.Ext(avatar.Filename)
	AvatarFields := map[string]bool{
		".jpg":  true,
		".png":  true,
		".jpeg": true,
		".webp": true,
	}
	if !AvatarFields[ext] {
		utils.BadRequest(c, "只支持 jpg, jpeg, png, webp 格式")
		return
	}

	uploadDir := "public/upload/avatars"
	timestamp := time.Now().Unix()
	fileName := fmt.Sprintf("user_%d_%d%s", userID, timestamp, ext)
	filePath := filepath.Join(uploadDir, fileName)
	if err := c.SaveUploadedFile(avatar, filePath); err != nil {
		utils.InternalServerError(c, "保存文件失败", err)
		return
	}
	avatarURL := fmt.Sprintf("http://localhost:8080/upload/avatars/%s", fileName)
	var user models.User
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		utils.BadRequest(c, "用户不存在")
		return
	}
	if user.Avatar != "" {
		oldFileName := filepath.Base(user.Avatar)
		oldFilePath := filepath.Join(uploadDir, oldFileName)
		if err := os.Remove(oldFilePath); err != nil {
			utils.BadRequest(c, "删除旧头像失败")
		}
	}
	if err := database.DB.Model(&user).Update("Avatar", avatarURL).Error; err != nil {
		utils.BadRequest(c, "更新头像失败")
		return
	}
	utils.SuccessWithMessage(c, "更新头像成功", gin.H{})
}

// 获取用户收藏
func GetUserCollection(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.BadRequest(c, "用户未鉴权")
		return
	}
	var teamIDs []string
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
		Offset((page-1)*pageSize).
		Limit(pageSize).
		Pluck("team_id", &teamIDs).Error; err != nil {
		utils.BadRequest(c, "查询收藏列表失败")
		return
	}
	utils.Success(c, gin.H{
		"total": total,
		"list":  teamIDs,
	})
}

// 添加收藏
func AddUserCollection(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.BadRequest(c, "用户未鉴权")
		return
	}

	teamID := c.Param("team_id")
	if teamID == "" {
		utils.BadRequest(c, "teamID不能为空")
		return
	}

	var team models.Team
	if err := database.DB.Where("team_id = ?", teamID).First(&team).Error; err != nil {
		utils.BadRequest(c, "队伍不存在")
		return
	}
	collections := models.UserCollection{
		UserID: userID,
		TeamID: teamID,
	}
	if err := database.GetDB().Create(&collections).Error; err != nil {
		utils.InternalServerError(c, "收藏失败:", err)
		return
	}
	utils.SuccessWithMessage(c, "收藏成功", collections)

}

// 移除收藏
func RemoveUserCollection(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.BadRequest(c, "用户未鉴权")
		return
	}

	teamID := c.Param("team_id")
	if teamID == "" {
		utils.BadRequest(c, "teamID不能为空")
		return
	}
	var team models.Team
	if err := database.DB.Where("team_id = ?", teamID).First(&team).Error; err != nil {
		utils.BadRequest(c, "队伍不存在")
		return
	}
	collections := models.UserCollection{
		UserID: userID,
		TeamID: teamID,
	}
	if err := database.GetDB().Unscoped().Delete(&collections).Error; err != nil {
		utils.InternalServerError(c, "移除收藏失败", err)
	}
	utils.SuccessWithMessage(c, "移除成功", collections)
}

// 上传简历
func UploadResume(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.BadRequest(c, "用户未鉴权")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "上传失败")
		return
	}
	if file.Size > 4*1024*1024 {
		utils.BadRequest(c, "上传文件应小于4MB")
		return
	}
	ext := filepath.Ext(file.Filename)
	ResumeFields := map[string]bool{
		".jpg":  true,
		".png":  true,
		".jpeg": true,
		".webp": true,
		".pdf":  true,
		".doc":  true,
		".docx": true,
	}
	if !ResumeFields[ext] {
		utils.BadRequest(c, "上传文件格式不符合要求")
		return
	}
	uploadDir := "public/upload/resumes"
	ResumeName := c.PostForm("fileName")
	filePath := filepath.Join(uploadDir, ResumeName)
	resume := models.UserResume{
		UserID:     userID,
		ResumeID:   uuid.New().String(),
		ResumeName: ResumeName,
		FilePath:   "/upload/resumes/" + ResumeName,
		FileType:   ext,
	}

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		utils.BadRequest(c, "文件保存失败")
		return
	}

	if err := database.DB.Create(&resume).Error; err != nil {
		utils.BadRequest(c, "上传简历失败")
		return
	}
	utils.SuccessWithMessage(c, "上传简历成功", gin.H{"resume": resume})
}

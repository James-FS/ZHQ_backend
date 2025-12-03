package controllers

import (
	"zhq-backend/database"
	"zhq-backend/models"
	"zhq-backend/utils"

	"github.com/gin-gonic/gin"
)

// GetFriendList 获取好友列表
func GetFriendList(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.BadRequest(c, "用户未鉴权")
		return
	}

	//只查正向的已通过记录，避免重复（因为双向记录都存在）
	var friends []models.Friend
	if err := database.GetDB().
		Where("user_id = ? AND status = 1", userID).
		Find(&friends).Error; err != nil {
		utils.InternalServerError(c, "获取好友列表失败:", err)
		return
	}

	//提取好友ID列表
	var friendIDs []string
	for _, f := range friends {
		friendIDs = append(friendIDs, f.FriendID)
	}

	//查询好友用户信息
	var users []models.User
	if len(friendIDs) > 0 {
		if err := database.GetDB().
			Where("user_id IN (?) AND user_id != ?", friendIDs, userID).
			Find(&users).Error; err != nil {
			utils.InternalServerError(c, "获取好友用户信息失败:", err)
			return
		}
	}

	utils.Success(c, gin.H{
		"list":  users,
		"count": len(users),
	})
}

// AddFriend 发送好友请求
func AddFriend(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.BadRequest(c, "用户未鉴权")
		return
	}

	var req struct {
		FriendID string `json:"friend_id" binding:"required"`
	}
	if err := c.ShouldBind(&req); err != nil {
		utils.BadRequest(c, "请求参数错误:"+err.Error())
		return
	}

	// 1.避免自己添加自己
	if userID == req.FriendID {
		utils.BadRequest(c, "不能添加自己为好友")
		return
	}

	// 2.检查好友是否存在
	var friendUser models.User
	if err := database.GetDB().
		Where("user_id = ?", req.FriendID).
		First(&friendUser).Error; err != nil {
		utils.BadRequest(c, "好友不存在")
		return
	}

	// 3.检查是否已发送请求或已是好友
	db := database.GetDB()

	// 检查正向请求（当前用户 -> 目标用户）
	var forwardRequest models.Friend
	forwardErr := db.Where("user_id = ? AND friend_id = ?", userID, req.FriendID).First(&forwardRequest).Error
	if forwardErr == nil {
		// 正向记录存在，任何状态都说明已发过请求
		utils.BadRequest(c, "好友请求已发送，正在等待对方验证")
		return
	}

	// 检查反向已通过（目标用户 -> 当前用户 且 status=1）
	var reverseAccepted models.Friend
	reverseErr := db.Where("user_id = ? AND friend_id = ? AND status = 1", req.FriendID, userID).First(&reverseAccepted).Error
	if reverseErr == nil {
		// 反向已通过，说明已是好友
		utils.BadRequest(c, "对方已是你的好友，无需重复添加")
		return
	}

	// 4.创建好友请求（正向记录，待对方通过后再创建反向记录）
	friend := models.Friend{
		UserID:   userID,
		FriendID: req.FriendID,
		Status:   0, //待验证
	}

	if err := db.Create(&friend).Error; err != nil {
		utils.InternalServerError(c, "发送好友请求失败:", err)
		return
	}

	utils.SuccessWithMessage(c, "好友请求发送成功，等待对方验证", friend)
}

// AcceptFriend 接受好友请求
func AcceptFriend(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.BadRequest(c, "用户未鉴权")
		return
	}

	friendID := c.Query("friend_id")
	if friendID == "" {
		utils.BadRequest(c, "friend_id不能为空")
		return
	}

	// 禁止接受自己的请求
	if userID == friendID {
		utils.BadRequest(c, "不能接受自己的好友请求")
		return
	}

	db := database.GetDB()
	// 1.开启数据库事务
	tx := db.Begin()
	defer func() {
		//处理panic异常：回滚事务
		if r := recover(); r != nil {
			tx.Rollback()
			errMsg := "处理好友请求时发生异常"
			if e, ok := r.(error); ok {
				errMsg += "：" + e.Error()
			}
			utils.InternalServerError(c, errMsg, nil)
		}
	}()

	// 2.幂等性校验（避免重复接受）
	var existingFriend models.Friend
	if err := tx.
		Where("user_id = ? AND friend_id = ? AND status = 1", friendID, userID).
		First(&existingFriend).Error; err == nil {
		tx.Rollback()
		utils.BadRequest(c, "已接受对方好友请求，无需重复操作")
		return
	}

	//3.检查请求是否存在（必须是待验证状态 status=0）
	if err := tx.
		Where("user_id = ? AND friend_id = ? AND status = 0", friendID, userID).
		First(&existingFriend).Error; err != nil {
		tx.Rollback()
		utils.BadRequest(c, "好友请求不存在或已处理")
		return
	}

	// 4. 事务内操作1：更新好友请求状态为已通过
	if err := tx.
		Model(&models.Friend{}).
		Where("user_id = ? AND friend_id = ? AND status = 0", friendID, userID).
		Update("status", 1).Error; err != nil {
		tx.Rollback()
		utils.InternalServerError(c, "接受好友请求失败", err)
		return
	}

	// 5. 事务内操作2：创建反向好友关系（自己→对方，直接已通过）
	reverseFriend := models.Friend{
		UserID:   userID,
		FriendID: friendID,
		Status:   1, // 已通过
	}
	if err := tx.Create(&reverseFriend).Error; err != nil {
		tx.Rollback()
		utils.InternalServerError(c, "添加好友失败", err)
		return
	}

	// 6. 所有操作成功，提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback() // 提交失败也回滚
		utils.InternalServerError(c, "提交好友关系失败", err)
		return
	}

	utils.SuccessWithMessage(c, "已成功添加好友", nil)

}

// RejectFriend 拒绝好友请求（完善幂等性和边界校验）
func RejectFriend(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.BadRequest(c, "用户未鉴权")
		return
	}

	friendID := c.Query("friend_id")
	if friendID == "" {
		utils.BadRequest(c, "friend_id不能为空")
		return
	}

	// 禁止拒绝自己的请求
	if userID == friendID {
		utils.BadRequest(c, "不能拒绝自己的好友请求")
		return
	}

	db := database.GetDB()

	// 新增：幂等性校验（避免重复拒绝）
	var existingFriend models.Friend
	if err := db.
		Where("user_id = ? AND friend_id = ? AND status = 2", friendID, userID).
		First(&existingFriend).Error; err == nil {
		utils.BadRequest(c, "已拒绝对方好友请求，无需重复操作")
		return
	}

	// 新增：检查请求是否存在（避免拒绝不存在的请求）
	if err := db.
		Where("user_id = ? AND friend_id = ? AND status = 0", friendID, userID).
		First(&existingFriend).Error; err != nil {
		utils.BadRequest(c, "好友请求不存在或已处理")
		return
	}

	// 更新好友状态为已拒绝
	if err := db.
		Model(&models.Friend{}).
		Where("user_id = ? AND friend_id = ?", friendID, userID).
		Update("status", 2).Error; err != nil {
		utils.InternalServerError(c, "拒绝好友请求失败", err)
		return
	}

	utils.SuccessWithMessage(c, "已拒绝好友请求", nil)
}

// GetFriendRequests 获取好友请求列表（待处理的请求）
func GetFriendRequests(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.BadRequest(c, "用户未鉴权")
		return
	}

	// 查询：对方→自己的待验证请求（status=0）
	var requests []models.Friend
	db := database.GetDB()
	if err := db.
		Where("friend_id = ? AND status = 0", userID).
		Find(&requests).Error; err != nil {
		utils.InternalServerError(c, "获取好友请求失败", err)
		return
	}

	// 提取请求者ID列表
	var requesterIDs []string
	for _, r := range requests {
		requesterIDs = append(requesterIDs, r.UserID)
	}

	// 查询请求者信息（排除自己，避免极端情况）
	var users []models.User
	if len(requesterIDs) > 0 {
		if err := db.
			Where("user_id IN (?) AND user_id != ?", requesterIDs, userID).
			Find(&users).Error; err != nil {
			utils.InternalServerError(c, "获取请求者信息失败", err)
			return
		}
	}

	utils.Success(c, gin.H{
		"list":  users,
		"count": len(users), // 新增：返回请求数量，方便前端展示
	})
}

// CheckFriendship 判断两个用户是否为好友
func CheckFriendship(c *gin.Context) {
	//获取当前登录用户ID
	currentUserID := c.GetString("user_id")
	if currentUserID == "" {
		utils.BadRequest(c, "用户未鉴权")
		return
	}

	//目标用户ID
	targetUserID := c.Query("target_user_id")
	if targetUserID == "" {
		utils.BadRequest(c, "target_user_id不能为空")
		return
	}

	//检查是否为同一用户
	if currentUserID == targetUserID {
		utils.BadRequest(c, "不能查询自己与自己的关系")
		return
	}

	//查询查询条件：双向记录且状态为已通过
	var count int64
	db := database.GetDB()
	err := db.Model(&models.Friend{}).
		Where("(user_id = ? AND friend_id = ? AND status = 1) OR (user_id = ? AND friend_id = ? AND status = 1)",
			currentUserID, targetUserID, //正向关系：当前用户 -> 目标用户
			targetUserID, currentUserID, //反向关系：目标用户 -> 当前用户
		).
		Count(&count).
		Error

	if err != nil {
		utils.InternalServerError(c, "查询好友关系失败:", err)
		return
	}

	//若双向记录存在，count=2，说明是好友
	isFriend := count == 2

	utils.Success(c, gin.H{
		"is_friend": isFriend,
		"user_id":   currentUserID,
		"target_id": targetUserID,
	})

}

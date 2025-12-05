package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"zhq-backend/database"
	"zhq-backend/models"
	"zhq-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket 升级器配置
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 生产环境应该检查具体的 Origin
		// 这里暂时允许所有来源
		return true
	},
}

// WebSocketHandler 处理 WebSocket 连接
// 这是 WebSocket 的入口点
func WebSocketHandler(c *gin.Context) {
	// 第 1 步：从上下文中获取用户 ID
	userID := c.GetString("user_id")
	if userID == "" {
		utils.Unauthorized(c, "用户未登录")
		return
	}

	fmt.Printf("🔗 用户 %s 尝试连接 WebSocket\n", userID)

	// 第 2 步：升级 HTTP 连接为 WebSocket
	// HTTP 1.1 ↓↑ WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("❌ WebSocket 升级失败: %v\n", err)
		utils.BadRequest(c, "升级websocket连接失败")
		return
	}

	fmt.Printf("✅ 用户 %s WebSocket 升级成功\n", userID)

	// 创建客户端对象
	client := utils.NewWebSocketClient(userID, conn)
	// 注册客户端到管理器
	utils.WSManager.Register <- client
	// 这条线会触发 WSManager.Run() 中的 case client := <-wm.Register:
	go utils.SendOfflineMessages(userID)
	go client.ReadMessages()  // 监听客户端发来的消息
	go client.WriteMessages() // 发送消息给客户端
}

// GetChatHistory 获取聊天历史记录
func GetChatHistory(c *gin.Context) {
	userID := c.GetString("user_id")
	otherUserID := c.Query("other_user_id")

	// 获取分页参数
	CountStr := c.DefaultQuery("count", "15")
	PageStr := c.DefaultQuery("page", "0")

	count, err := strconv.Atoi(CountStr)
	if err != nil || count <= 0 {
		count = 15
	}

	page, err := strconv.Atoi(PageStr)
	if err != nil || page < 0 {
		page = 0
	}

	// 验证参数
	if otherUserID == "" {
		utils.BadRequest(c, "缺少 other_user_id 参数")
		return
	}

	// 获取或创建会话
	sessionID, err := utils.GetOrCreateSession(userID, otherUserID)
	if err != nil {
		utils.InternalServerError(c, "获取会话失败: ", err)
		return
	}

	// 从数据库查询消息（按时间升序排列）
	var messages []models.Message
	result := database.DB.
		Where("session_id = ?  AND deleted_at IS NULL", sessionID).
		Order("created_at ASC").
		Limit(count).
		Offset((page - 1) * count).
		Find(&messages)

	if result.Error != nil {
		utils.InternalServerError(c, "查询消息失败: ", result.Error)
		return
	}

	// 获取总消息数
	var total int64
	database.DB.
		Model(&models.Message{}).
		Where("session_id = ?  AND deleted_at IS NULL", sessionID).
		Count(&total)

	// 返回成功响应
	utils.Success(c, gin.H{
		"messages": messages,
		"total":    total,
		"limit":    count,
		"offset":   page,
	})
}

// GetChatList 获取聊天列表
func GetChatList(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.BadRequest(c, "用户未登录")
		return
	}
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	var sessions []models.ChatSession
	var total int64
	offset := (page - 1) * pageSize
	list := database.DB.
		Where("(user_id1 = ? OR user_id2 = ?) AND deleted_at IS NULL", userID, userID).
		Model(&models.ChatSession{}).
		Count(&total).
		Order("last_msg_time DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&sessions)

	if list.Error != nil {
		utils.InternalServerError(c, "查询聊天列表失败: ", list.Error)
		return
	}

	type ChatListItem struct {
		SessionID     string     `json:"session_id"`     // 会话ID
		LastMessage   string     `json:"last_message"`   // 最后一条消息
		LastMsgTime   *time.Time `json:"last_msg_time"`  // 最后消息时间
		OtherUserID   string     `json:"other_user_id"`  // 另一个用户ID
		SessionName   string     `json:"session_name"`   // 另一个用户昵称
		SessionAvatar string     `json:"session_avatar"` // 另一个用户头像
	}

	var chatList []ChatListItem
	for _, session := range sessions {
		var otherUserID string
		if session.UserID1 == userID {
			otherUserID = session.UserID2
		} else {
			otherUserID = session.UserID1
		}
		var otherUser models.User
		if err := database.DB.
			Where("user_id = ? ", otherUserID).
			First(&otherUser).Error; err != nil {
			fmt.Printf("❌ 查询用户信息失败: %v\n", err)
			continue
		}
		chatList = append(chatList, ChatListItem{
			SessionID:     session.SessionID,
			LastMessage:   session.LastMessage,
			LastMsgTime:   session.LastMsgTime,
			OtherUserID:   otherUserID,
			SessionName:   otherUser.Nickname,
			SessionAvatar: otherUser.Avatar,
		})
	}
	utils.Success(c, gin.H{
		"chat_list": chatList,
		"total":     total,
	})

}

// GetOnlineUsers 获取在线用户列表
func GetOnlineUsers(c *gin.Context) {
	users := utils.WSManager.GetOnlineUsers()
	utils.Success(c, gin.H{
		"online_users": users,
		"count":        len(users),
	})
}

// CheckUserOnline 检查指定用户是否在线
func CheckUserOnline(c *gin.Context) {
	targetUserID := c.Query("user_id")
	if targetUserID == "" {
		utils.BadRequest(c, "user_id 不能为空")
		return
	}

	isOnline := utils.WSManager.IsUserOnline(targetUserID)
	utils.Success(c, gin.H{
		"user_id":   targetUserID,
		"is_online": isOnline,
	})
}

package controllers

import (
	"fmt"
	"net/http"
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
	// 第 1 步：从上下文中获取用户 ID（从 JWT token 中解析出来）
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

	go client.ReadMessages()  // 监听客户端发来的消息
	go client.WriteMessages() // 发送消息给客户端
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

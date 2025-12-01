package utils

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"zhq-backend/database"
	"zhq-backend/models"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ============ 数据结构定义 ============

// MessagePayload WebSocket 消息的格式
// 这是客户端发来的消息，或者服务器发给客户端的消息
type MessagePayload struct {
	Type      string                 `json:"type"`           // 消息类型：chat, typing, online 等
	SenderID  string                 `json:"sender_id"`      // 发送者 ID
	Content   string                 `json:"content"`        // 消息内容
	Timestamp time.Time              `json:"timestamp"`      // 消息时间戳
	Data      map[string]interface{} `json:"data,omitempty"` // 额外数据
}

// WebSocketClient 代表一个连接的客户端
type WebSocketClient struct {
	UserID   string           // 用户 ID
	Conn     *websocket.Conn  // WebSocket 连接对象
	Send     chan interface{} // 消息发送队列（缓冲区大小 256）
	IsActive bool             // 连接是否活跃
	mu       sync.RWMutex     // 读写锁，保护并发访问
}

// WebSocketManager 管理所有 WebSocket 连接
type WebSocketManager struct {
	Clients    map[string]*WebSocketClient // 存储所有已连接的客户端 map[userID]client
	Broadcast  chan interface{}            // 广播消息通道
	Register   chan *WebSocketClient       // 新客户端注册通道
	Unregister chan *WebSocketClient       // 客户端注销通道
	mu         sync.RWMutex                // 保护 Clients map 的并发访问
}

// 全局的 WebSocket 管理器实例
var WSManager = &WebSocketManager{
	Clients:    make(map[string]*WebSocketClient),
	Broadcast:  make(chan interface{}, 256),
	Register:   make(chan *WebSocketClient),
	Unregister: make(chan *WebSocketClient),
}

// ============ 初始化和启动 ============
func (wm *WebSocketManager) Init() {
	go wm.Run()
	fmt.Println("✅ WebSocket 管理器已初始化")
}

func (wm *WebSocketManager) Run() {
	for {
		select {
		// 监听新客户端注册
		case client := <-wm.Register:
			wm.mu.Lock()
			wm.Clients[client.UserID] = client
			wm.mu.Unlock()
			fmt.Printf("✅ 用户 %s 已连接，当前在线人数: %d\n", client.UserID, len(wm.Clients))

			// 发送连接成功消息给该用户
			wm.SendToUser(client.UserID, map[string]interface{}{
				"type":    "connected",
				"status":  "success",
				"user_id": client.UserID,
			})

		// 监听客户端注销
		case client := <-wm.Unregister:
			wm.mu.Lock()
			if _, exists := wm.Clients[client.UserID]; exists {
				delete(wm.Clients, client.UserID)
				close(client.Send)
			}
			wm.mu.Unlock()
			fmt.Printf("❌ 用户 %s 已断开连接，当前在线人数: %d\n", client.UserID, len(wm.Clients))

		// 监听广播消息
		case message := <-wm.Broadcast:
			wm.mu.RLock()
			for _, client := range wm.Clients {
				select {
				case client.Send <- message:
				default:
					// 发送缓冲区已满，跳过
					fmt.Printf("⚠️ 无法发送广播消息给用户 %s（缓冲区已满）\n", client.UserID)
				}
			}
			wm.mu.RUnlock()
		}
	}
}

// GetOrCreateSession 获取或创建会话 ID
func GetOrCreateSession(userID1, userID2 string) (string, error) {
	// 确保 userID1 < userID2 以保持一致性
	if userID1 > userID2 {
		userID1, userID2 = userID2, userID1
	}

	// 查找是否已存在该会话
	var existingSession models.ChatSession
	result := database.DB.Where("user_id_1 = ? AND user_id_2 = ?  AND session_type = ?", userID1, userID2, "private").
		First(&existingSession)

	if result.Error == nil {
		// 会话已存在
		return existingSession.SessionID, nil
	}

	// 创建新会话
	session := models.ChatSession{
		SessionID:   uuid.New().String(),
		SessionType: "private",
		UserID1:     userID1,
		UserID2:     userID2,
	}

	if err := database.DB.Create(&session).Error; err != nil {
		return "", fmt.Errorf("创建会话失败: %v", err)
	}

	return session.SessionID, nil
}

func SaveMessageToDatabase(receiverID string, message interface{}) error {
	// 将消息转换为 MessagePayload
	var payload MessagePayload

	if msgPayload, ok := message.(MessagePayload); ok {
		payload = msgPayload
	} else {
		// 如果是 map 格式，转换为 MessagePayload
		msgJSON, _ := json.Marshal(message)
		json.Unmarshal(msgJSON, &payload)
	}

	// 如果消息类型不是 chat，则不保存
	if payload.Type != "chat" {
		return nil
	}

	var sessionID string
	if sid, ok := payload.Data["session_id"].(string); ok {
		sessionID = sid
	}
	// 创建消息对象
	msg := models.Message{
		MessageID:    uuid.New().String(),
		SenderID:     payload.SenderID,
		ReceiverID:   receiverID,
		SessionID:    sessionID,
		Content:      payload.Content,
		ContentTypes: "text",
		SendStatus:   false,
	}

	// 保存到数据库
	if err := database.DB.Create(&msg).Error; err != nil {
		return fmt.Errorf("保存离线消息失败: %v", err)
	}

	fmt.Printf("💾 离线消息已保存: %s -> %s\n", payload.SenderID, receiverID)
	return nil
}

// SendOfflineMessages 发送用户的所有离线消息
func SendOfflineMessages(userID string) {
	// 查询所有发送给该用户的消息
	var messages []models.Message
	if err := database.DB.Where("receiver_id = ? AND send_status = ?", userID, false).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		fmt.Printf("❌ 查询离线消息失败: %v\n", err)
		return
	}

	if len(messages) == 0 {
		return
	}

	count := 0
	for _, msg := range messages {
		// 构造 MessagePayload 格式发送
		payload := MessagePayload{
			Type:      "chat",
			SenderID:  msg.SenderID,
			Content:   msg.Content,
			Timestamp: msg.UpdatedAt,
			Data: map[string]interface{}{
				"receiver_id": msg.ReceiverID,
				"session_id":  msg.SessionID,
				"message_id":  msg.MessageID,
			},
		}

		// 检查用户是否已连接
		WSManager.mu.RLock()
		client, exists := WSManager.Clients[userID]
		WSManager.mu.RUnlock()

		if exists && client.IsActive {
			select {
			case client.Send <- payload:
				count++
				// 发送成功后，更新消息状态为已发送
				if err := database.DB.Model(&msg).Update("send_status", true).Error; err != nil {
					fmt.Printf("❌ 更新消息状态失败: %v\n", err)
				}
			default:
				fmt.Printf("⚠️ 无法发送消息（缓冲区已满）: %s\n", msg.MessageID)
			}
		}
	}

	if count > 0 {
		fmt.Printf("📬 用户 %s 收到 %d 条离线消息\n", userID, count)
	}
}

// ============ 客户端管理 ============
// NewWebSocketClient 创建新的 WebSocket 客户端
func NewWebSocketClient(userID string, conn *websocket.Conn) *WebSocketClient {
	return &WebSocketClient{
		UserID:   userID,
		Conn:     conn,
		Send:     make(chan interface{}, 256), // 256 大小的缓冲队列
		IsActive: true,
	}
}

// SendToUser 发送消息到指定用户

func (wm *WebSocketManager) SendToUser(userID string, message interface{}) error {
	wm.mu.RLock()
	client, exists := wm.Clients[userID]
	wm.mu.RUnlock()

	if !exists || !client.IsActive {
		return SaveMessageToDatabase(userID, message)
	}

	select {
	case client.Send <- message:
		return nil
	default:
		return SaveMessageToDatabase(userID, message)
	}
}

// GetOnlineUsers 获取所有在线用户列表
// 返回值：
//
//	[]string - 在线用户的 ID 列表
func (wm *WebSocketManager) GetOnlineUsers() []string {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	var users []string
	for userID := range wm.Clients {
		users = append(users, userID)
	}
	return users
}

// IsUserOnline 检查用户是否在线
func (wm *WebSocketManager) IsUserOnline(userID string) bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	_, exists := wm.Clients[userID]
	return exists
}

// ============ 客户端读写消息 ============

// ReadMessages 持续读取客户端发送的消息
// 这是一个 goroutine，会一直运行直到连接断开
func (c *WebSocketClient) ReadMessages() {
	defer func() {
		// 连接关闭时的清理工作
		c.mu.Lock()
		c.IsActive = false
		c.mu.Unlock()

		// 注销这个客户端
		WSManager.Unregister <- c
		c.Conn.Close()
	}()

	// 设置读超时：60 秒内必须收到消息或 ping/pong，否则连接超时
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// 设置心跳响应处理
	// 当客户端回复 pong 时，重置超时时间
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		// 阻塞等待客户端发送消息
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			// 检查是否是正常关闭或异常关闭
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("❌ WebSocket 错误 [%s]: %v\n", c.UserID, err)
			}
			break
		}

		// 解析接收到的 JSON 消息
		var payload MessagePayload
		if err := json.Unmarshal(data, &payload); err != nil {
			fmt.Printf("❌ 消息解析错误 [%s]: %v\n", c.UserID, err)
			continue
		}

		// 填充发送者信息和时间戳
		payload.SenderID = c.UserID
		payload.Timestamp = time.Now()

		// 根据消息类型处理不同的逻辑
		switch payload.Type {
		case "chat":
			// 一对一聊天消息 - 发送给指定接收者
			if receiver, ok := payload.Data["receiver_id"].(string); ok {
				fmt.Printf("💬 消息转发: %s -> %s\n", c.UserID, receiver)
				sessionID, err := GetOrCreateSession(c.UserID, receiver)
				if err != nil {
					fmt.Printf("❌ 获取会话失败: %v\n", err)
					continue
				}
				payload.Data["session_id"] = sessionID
				if err := WSManager.SendToUser(receiver, payload); err != nil {
					fmt.Printf("⚠️ 消息发送失败: %v\n", err)
				}
			}

		//case "typing":
		//	// 正在输入提示 - 发送给指定接收者
		//	if receiver, ok := payload.Data["receiver_id"].(string); ok {
		//		fmt.Printf("⌨️ 输入提示: %s 正在输入\n", c.UserID)
		//		WSManager.SendToUser(receiver, payload)
		//	}

		case "online":
			// 用户在线状态更新 - 广播给所有用户
			fmt.Printf("🟢 用户在线: %s\n", c.UserID)
			WSManager.Broadcast <- payload

		default:
			fmt.Printf("❓ 未知消息类型: %s\n", payload.Type)
		}
	}
}

// WriteMessages 持续发送消息给客户端
// 这是一个 goroutine，会一直运行直到连接断开
func (c *WebSocketClient) WriteMessages() {
	// 创建一个 54 秒的心跳定时器
	// 每 54 秒发送一次 ping 给客户端，客户端会回复 pong
	ticker := time.NewTicker(54 * time.Second)

	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		// 监听是否有消息要发送给客户端
		case message, ok := <-c.Send:
			// 设置写超时：10 秒内必须完成写入
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			if !ok {
				// 通道已关闭，连接已断开
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 以 JSON 格式发送消息给客户端
			if err := c.Conn.WriteJSON(message); err != nil {
				fmt.Printf("❌ 消息发送错误 [%s]: %v\n", c.UserID, err)
				return
			}

		// 监听心跳定时器触发（每 54 秒）
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			// 发送 ping 给客户端
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Printf("❌ 心跳发送错误 [%s]: %v\n", c.UserID, err)
				return
			}
		}
	}
}

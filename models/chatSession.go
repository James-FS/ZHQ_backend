package models

import "time"

type ChatSession struct {
	BaseModel
	SessionID    string     `json:"session_id" gorm:"uniqueIndex;not null comment:对话ID"` // 会话ID（支持群组）
	SessionType  string     `json:"session_type" gorm:"default:'private'"`               // private 一对一, group 群组
	LastMessage  string     `json:"last_message"`                                        // 最后一条消息内容
	LastMsgTime  *time.Time `json:"last_msg_time"`                                       // 最后消息时间
	LastSenderID string     `json:"last_sender_id"`                                      // 最后发送者ID
	MessageCount int64      `json:"message_count" gorm:"default:0"`                      // 消息总数
	UserID1      string     `json:"user_id_1" gorm:"index;not null"`                     // 用户1
	UserID2      string     `json:"user_id_2" gorm:"index;not null"`                     // 用户2
	User1        User       `json:"user1" gorm:"foreignKey:UserID1;references:UserID"`
	User2        User       `json:"user2" gorm:"foreignKey:UserID2;references:UserID"`
	//Messages     []Message  `json:"messages" gorm:"foreignKey:SessionID"`
}

// TableName 设置表名
func (ChatSession) TableName() string {
	return "chat_sessions"
}

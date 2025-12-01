package models

type Message struct {
	BaseModel
	MessageID    string `json:"message_id" gorm:"uniqueIndex;not null comment:消息ID"` // UUID
	SenderID     string `json:"sender_id" gorm:"index;not null comment:发送者ID"`       // 发送者 user_id
	ReceiverID   string `json:"receiver_id" gorm:"index;not null comment:接收者ID"`     // 接收者 user_id（一对一时）
	SessionID    string `json:"session_id" gorm:"index;not null comment:对话ID"`       // 会话ID（支持群组）
	Content      string `json:"content" gorm:"type:longtext;not null comment:内容"`    // 富文本内容
	ContentTypes string `json:"content_type" gorm:"default:'text' comment:内容类型"`     // text, image, file, richtext
	SendStatus   bool   `json:"send_status" gorm:"comment:发送状态"`
	Sender       User   `json:"sender" gorm:"foreignKey:SenderID;references:UserID"`
	Receiver     User   `json:"receiver" gorm:"foreignKey:ReceiverID;references:UserID"`
}

// TableName 设置表名
func (Message) TableName() string {
	return "message"
}

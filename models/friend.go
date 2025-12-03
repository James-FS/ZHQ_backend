package models

// Friend 好友关系表
type Friend struct {
	BaseModel
	UserID   string `json:"user_id" gorm:"type:varchar(64);not null;index:idx_user_friend,unique"`   // 用户ID
	FriendID string `json:"friend_id" gorm:"type:varchar(64);not null;index:idx_user_friend,unique"` // 好友用户ID
	Status   int    `json:"status" gorm:"default:0;comment:0-待验证 1-已通过 2-已拒绝"`                       // 状态字段
}

func (Friend) TableName() string {
	return "friends"
}

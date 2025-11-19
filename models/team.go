package models

type Team struct {
	BaseModel
	TeamID         string `json:"team_id" gorm:"type:varchar(64);uniqueIndex;not null;comment:队伍ID(对外唯一业务标识)"`
	TeamName       string `json:"team_name" gorm:"type:varchar(100);not null;comment:队伍名称"`
	Description    string `json:"description" gorm:"type:text;comment:队伍项目内容"`
	Pictures       string `json:"pictures" gorm:"type:text;comment:队伍项目图片，多个图片URL以逗号分隔"`
	CreatorID      uint   `json:"creator_id" gorm:"not null;comment:创建者用户ID"`
	Status         int    `json:"status" gorm:"default:1;comment:状态 1:招募中 2:已截止 3:开发中 4:已完成"`
	MaxMembers     int    `json:"max_members" gorm:"default:5;comment:最大成员数"`
	CurrentMembers int    `json:"current_members" gorm:"default:1;comment:当前成员数"`
	Tags           string `json:"tags" gorm:"type:varchar(255);comment:标签，多个标签以逗号分隔"`
}

// TableName 设置表名
func (Team) TableName() string {
	return "teams"
}

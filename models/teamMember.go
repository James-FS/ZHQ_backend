package models

// TeamMember 队伍成员关系表
type TeamMember struct {
	BaseModel
	TeamID string `json:"team_id" gorm:"type:varchar(64);not null;index: idx_team_user;comment:队伍ID"`
	UserID string `json:"user_id" gorm:"type:varchar(64);not null;index:idx_team_user;index:idx_user;comment:用户ID"`
	Role   int    `json:"role" gorm:"type:tinyint;default:0;comment:角色 0:普通成员 1:队长"`
}

// TableName 设置表名
func (TeamMember) TableName() string {
	return "team_members"
}

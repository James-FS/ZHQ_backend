package models

type UserCollection struct {
	BaseModel
	TeamID string `json:"team_id" gorm:"type:varchar(64);not null;comment:队伍ID;index:idx_user_team,unique"`
	UserID string `json:"user_id" gorm:"type:varchar(64);not null;comment:用户业务ID;index:idx_user_team,unique"`
}

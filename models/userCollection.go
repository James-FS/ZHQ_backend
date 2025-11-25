package models

import "time"

type UserCollection struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	TeamID    string    `json:"team_id" gorm:"type:varchar(64);not null;comment:队伍ID;index:idx_user_team,unique"`
	UserID    string    `json:"user_id" gorm:"type:varchar(64);not null;comment:用户业务ID;index:idx_user_team,unique"`
}

func (UserCollection) TableName() string {
	return "user_collections"
}

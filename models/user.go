package models

type User struct {
	BaseModel
	OpenID   string `json:"openid" gorm:"type:varchar(191);uniqueIndex;not null;comment:微信OpenID"`
	UnionID  string `json:"unionid" gorm:"type:varchar(191);index;comment:微信UnionID"`
	Nickname string `json:"nickname" gorm:"type:varchar(100);comment:昵称"`
	Avatar   string `json:"avatar" gorm:"type:varchar(500);comment:头像URL"`
	Gender   int    `json:"gender" gorm:"default:0;comment:性别 0:未知 1:男 2:女"`
	Phone    string `json:"phone" gorm:"type:varchar(20);index;comment:手机号"`
	Status   int    `json:"status" gorm:"default:1;comment:状态 1:正常 0:禁用"`
}

// TableName 设置表名
func (User) TableName() string {
	return "users"
}

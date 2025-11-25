package models

type User struct {
	BaseModel
	UserID   string `json:"user_id" gorm:"type:varchar(64);uniqueIndex;not null;comment:用户业务ID(对外唯一标识)"`
	OpenID   string `json:"openid" gorm:"type:varchar(191);uniqueIndex;not null;comment:微信OpenID"`
	UnionID  string `json:"unionid" gorm:"type:varchar(191);index;comment:微信UnionID"`
	Nickname string `json:"nickname" gorm:"type:varchar(100);comment:昵称"`
	Avatar   string `json:"avatar" gorm:"type:varchar(500);comment:头像URL"`
	Gender   int    `json:"gender" gorm:"default:0;comment:性别 0:未知 1:男 2:女"`
	Phone    string `json:"phone" gorm:"type:varchar(20);index;comment:手机号"`
	College  string `json:"college" gorm:"type:varchar(64);comment:学院"`
	major    string `json:"major" gorm:"type:varchar(64);comment:专业"`
	Status   int    `json:"status" gorm:"default:1;comment:状态 1:正常 0:禁用"`
	Tags     string `json:"tags" gorm:"type:json;comment:标签数组（JSON格式字符串）"`
}

// TableName 设置表名
func (User) TableName() string {
	return "users"
}

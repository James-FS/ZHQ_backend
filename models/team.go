package models

type Team struct {
	BaseModel
	Name    string   `json:"name" gorm:"index"`
	Leader  uint     `json:"leader" gorm:"type:uint;index"`
	Members []uint   `json:"members" gorm:"type:json"`
	Status  string   `json:"status" gorm:"type:varchar;index"`
	Tags    []string `json:"tags" gorm:"type:json;index"`
	Intro   string   `json:"intro" gorm:"type:varchar;index"`
}

func (Team) TableName() string {
	return "Team"
}

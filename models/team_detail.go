package models

type TeamDetail struct {
	BaseModel
	Content             string `gorm:"type:text"`
	AnticipativeOutcome string `json:"anticipative_outcome" gorm:"type:text"`
	Recruitment         uint   `json:"recruitment" gorm:"type:uint"`
	Requirement         string `json:"requirement" gorm:"type:text"`
	Period              string `json:"period" gorm:"type:varchar"`
	Contest             string `json:"contest" gorm:"type:text"`
}

func (TeamDetail) TableName() string { return "TeamDetail" }

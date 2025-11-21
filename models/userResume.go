package models

type UserResume struct {
	BaseModel
	UserID     string `json:"user_id" gorm:"type:varchar(64);not null;comment:用户业务ID;"`
	ResumeID   string `json:"resume_id gorm:type:varchar(64);not null;comment:简历ID"`
	ResumeName string `json:"resume_name gorm:type:varchar(255);not null comment:简历名"`
	FilePath   string `json:"file_path gorm:type:text;not null comment:文件路径"`
	FileType   string `json:"file_type gorm:type:varchar(64);not null comment:文件类型"`
}

func (UserResume) TableName() string {
	return "UserResume"
}

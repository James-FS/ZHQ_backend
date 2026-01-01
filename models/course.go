package models

import "time"

// Course 课程表模型
type Course struct {
	BaseModel
	UserID       string `json:"user_id" gorm:"type:varchar(64);not null;index;comment: 用户ID"`
	CourseName   string `json:"course_name" gorm:"type:varchar(200);not null;comment:课程名称"`               // ← 增加到200
	Teacher      string `json:"teacher" gorm:"type:varchar(100);comment:教师姓名"`                            // ← 增加到100
	Classroom    string `json:"classroom" gorm:"type: varchar(150);comment:教室"`                           // ← 增加到150
	WeekDay      int    `json:"week_day" gorm:"type:int;not null;comment:星期几(1-7)"`                       // ← 添加not null
	StartWeek    int    `json:"start_week" gorm:"type:int;default:1;comment:开始周"`                         // ← 添加默认值
	EndWeek      int    `json:"end_week" gorm:"type:int;default:18;comment:结束周"`                          // ← 添加默认值
	StartSection int    `json:"start_section" gorm:"type:int;default:1;comment:开始节次"`                     // ← 添加默认值
	EndSection   int    `json:"end_section" gorm:"type:int;default:2;comment: 结束节次"`                      // ← 添加默认值
	Semester     string `json:"semester" gorm:"type:varchar(30);not null;index;comment:学期"`               // ← 增加到30，添加索引
	WeekType     string `json:"week_type" gorm:"type:varchar(20);default:'全周';comment:周类型(单周/双周/全周/指定周)"` // ← 添加默认值
}

func (Course) TableName() string {
	return "courses"
}

// CourseSchedulePDF 课程表PDF记录
type CourseSchedulePDF struct {
	BaseModel
	UserID     string     `json:"user_id" gorm:"type:varchar(64);not null;index;comment:用户ID"`
	FileName   string     `json:"file_name" gorm:"type:varchar(255);comment:文件名"`
	FilePath   string     `json:"file_path" gorm:"type: text;comment:文件路径"`
	Semester   string     `json:"semester" gorm:"type:varchar(30);comment:学期"` // ← 增加到30
	UploadTime *time.Time `json:"upload_time" gorm:"comment:上传时间"`
}

func (CourseSchedulePDF) TableName() string {
	return "course_schedule_pdfs"
}

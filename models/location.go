package models

// Location 校园地点信息
type Location struct {
	BaseModel
	Name      string  `json:"name" gorm:"index;comment:地点名称"`
	Category  string  `json:"category" gorm:"index;comment:地点类型"`
	Latitude  float64 `json:"latitude" gorm:"comment:纬度"`
	Longitude float64 `json:"longitude" gorm:"comment:经度"`
	Tags      string  `json:"tags" gorm:"comment:标签(JSON数组)"`
	Status    int     `json:"status" gorm:"default:1;comment:状态(1:有效,0:无效)"`
}

// TableName 指定表名
func (Location) TableName() string {
	return "locations"
}

// LocationCategory 地点类别常量
const (
	CategoryTeaching   = "teaching"   // 教学楼
	CategoryDorm       = "dorm"       // 宿舍
	CategoryDining     = "dining"     // 食堂
	CategorySports     = "sports"     // 体育设施
	CategoryLibrary    = "library"    // 图书馆
	CategoryLab        = "lab"        // 实验楼
	CategoryMedical    = "medical"    // 门诊
	CategoryCommercial = "commercial" // 商业中心
	CategoryOther      = "other"      // 其他
)

// LocationListResponse 地点列表响应
type LocationListResponse struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// LocationDetailResponse 地点详情响应
type LocationDetailResponse struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Tags      string  `json:"tags"`
}

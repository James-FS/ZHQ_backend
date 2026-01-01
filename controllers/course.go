package controllers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"zhq-backend/database"
	"zhq-backend/models"
	"zhq-backend/utils"

	"github.com/gin-gonic/gin"
)

// UploadCoursePDF 上传课程表PDF
func UploadCoursePDF(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.Unauthorized(c, "用户未登录")
		return
	}

	semester := c.PostForm("semester")
	if semester == "" {
		utils.BadRequest(c, "学期参数不能为空")
		return
	}

	// 接收PDF文件
	file, err := c.FormFile("pdf")
	if err != nil {
		utils.BadRequest(c, "请上传PDF文件")
		return
	}

	// 验证文件类型
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".pdf") {
		utils.BadRequest(c, "只支持PDF格式")
		return
	}

	// 保存文件
	uploadDir := "public/upload/course_schedules"
	os.MkdirAll(uploadDir, 0755)

	filename := fmt.Sprintf("%s_%d_%s", userID, time.Now().Unix(), file.Filename)
	filePath := filepath.Join(uploadDir, filename)

	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		utils.InternalServerError(c, "保存文件失败", err)
		return
	}

	// 解析PDF
	courses, err := utils.ParseCoursePDF(filePath)
	if err != nil {
		os.Remove(filePath)
		utils.InternalServerError(c, "解析PDF失败:  "+err.Error(), err)
		return
	}

	if len(courses) == 0 {
		os.Remove(filePath)
		utils.BadRequest(c, "未能从PDF中识别到课程信息，请检查PDF格式")
		return
	}

	db := database.GetDB()

	// 删除该学期的旧数据
	db.Where("user_id = ? AND semester = ?", userID, semester).Delete(&models.Course{})

	// 保存新数据
	for _, courseInfo := range courses {
		course := models.Course{
			UserID:       userID,
			CourseName:   courseInfo.CourseName,
			Teacher:      courseInfo.Teacher,
			Classroom:    courseInfo.Classroom,
			WeekDay:      courseInfo.WeekDay,
			StartWeek:    courseInfo.StartWeek,
			EndWeek:      courseInfo.EndWeek,
			StartSection: courseInfo.StartSection,
			EndSection:   courseInfo.EndSection,
			Semester:     semester,
			WeekType:     courseInfo.WeekType,
		}

		if err := db.Create(&course).Error; err != nil {
			utils.InternalServerError(c, "保存课程失败", err)
			return
		}
	}

	utils.Success(c, gin.H{
		"message": "导入成功",
		"count":   len(courses),
		"courses": courses,
	})
}

// UploadCourseHTML 上传HTML格式的课程表
// UploadCourseHTML 上传HTML格式的课程表
func UploadCourseHTML(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.Unauthorized(c, "用户未登录")
		return
	}

	semester := c.PostForm("semester")
	if semester == "" {
		utils.BadRequest(c, "学期参数不能为空")
		return
	}

	// 接收HTML文件
	file, err := c.FormFile("html")
	if err != nil {
		utils.BadRequest(c, "请上传HTML文件")
		return
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".html" && ext != ".htm" {
		utils.BadRequest(c, "只支持HTML格式")
		return
	}

	// 保存临时文件
	tempDir := "temp"
	os.MkdirAll(tempDir, 0755)
	tempPath := filepath.Join(tempDir, fmt.Sprintf("%s_%d_%s", userID, time.Now().Unix(), file.Filename))

	err = c.SaveUploadedFile(file, tempPath)
	if err != nil {
		utils.InternalServerError(c, "保存文件失败", err)
		return
	}
	defer os.Remove(tempPath)

	fmt.Printf("📁 [DEBUG] HTML文件保存到: %s\n", tempPath)

	// 直接解析HTML表格
	fmt.Printf("🔍 [DEBUG] 尝试解析HTML表格...\n")
	courses, err := utils.ParseCourseHTML(tempPath)
	if err != nil {
		utils.InternalServerError(c, "解析HTML失败:  "+err.Error(), err)
		return
	}

	// 如果没有识别到课程，回退到文本解析
	if len(courses) == 0 {
		fmt.Printf("⚠️ [DEBUG] 表格解析失败，回退到文本解析...\n")
		text, err := utils.ExtractTextFromHTMLFile(tempPath)
		if err != nil {
			utils.InternalServerError(c, "解析HTML失败:  "+err.Error(), err)
			return
		}

		if text == "" {
			utils.BadRequest(c, "HTML文件为空")
			return
		}

		fmt.Printf("✅ [DEBUG] 提取文本长度:  %d 字符\n", len(text))
		fmt.Printf("📝 [DEBUG] 文本预览:\n%s\n", text[:utils.Min(1000, len(text))])

		courses = utils.ParseCourseText(text)
	}

	fmt.Printf("🔍 [DEBUG] 解析到 %d 门课程\n", len(courses))

	if len(courses) == 0 {
		utils.BadRequest(c, "未能从HTML中识别到课程信息")
		return
	}

	// 删除旧数据
	db := database.GetDB()
	db.Where("user_id = ?   AND semester = ?", userID, semester).Delete(&models.Course{})

	// 保存新数据
	for _, courseInfo := range courses {
		course := models.Course{
			UserID:       userID,
			CourseName:   courseInfo.CourseName,
			Teacher:      courseInfo.Teacher,
			Classroom:    courseInfo.Classroom,
			WeekDay:      courseInfo.WeekDay,
			StartWeek:    courseInfo.StartWeek,
			EndWeek:      courseInfo.EndWeek,
			StartSection: courseInfo.StartSection,
			EndSection:   courseInfo.EndSection,
			Semester:     semester,
			WeekType:     courseInfo.WeekType,
		}

		if err := db.Create(&course).Error; err != nil {
			utils.InternalServerError(c, "保存课程失败", err)
			return
		}
	}

	utils.Success(c, gin.H{
		"message": "导入成功",
		"count":   len(courses),
		"courses": courses,
	})
}

// TestParseHTML 测试解析HTML（不保存到数据库）
func TestParseHTML(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.Unauthorized(c, "用户未登录")
		return
	}

	// 接收HTML文件
	file, err := c.FormFile("html")
	if err != nil {
		utils.BadRequest(c, "请上传HTML文件")
		return
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".html" && ext != ".htm" {
		utils.BadRequest(c, "只支持HTML格式")
		return
	}

	// 保存临时文件
	tempDir := "temp"
	os.MkdirAll(tempDir, 0755)
	tempPath := filepath.Join(tempDir, fmt.Sprintf("%s_%d_%s", userID, time.Now().Unix(), file.Filename))

	err = c.SaveUploadedFile(file, tempPath)
	if err != nil {
		utils.InternalServerError(c, "保存文件失败", err)
		return
	}
	defer os.Remove(tempPath)

	fmt.Printf("📁 [DEBUG] HTML文件保存到: %s\n", tempPath)

	// 直接解析HTML表格
	fmt.Printf("🔍 [DEBUG] 尝试解析HTML表格...\n")
	courses, err := utils.ParseCourseHTML(tempPath)
	if err != nil {
		utils.InternalServerError(c, "解析HTML失败:  "+err.Error(), err)
		return
	}

	// 如果没有识别到课程，回退到文本解析
	if len(courses) == 0 {
		fmt.Printf("⚠️ [DEBUG] 表格解析失败，回退到文本解析.. .\n")
		text, err := utils.ExtractTextFromHTMLFile(tempPath)
		if err != nil {
			utils.InternalServerError(c, "解析HTML失败: "+err.Error(), err)
			return
		}

		if text == "" {
			utils.BadRequest(c, "HTML文件为空")
			return
		}

		fmt.Printf("✅ [DEBUG] 提取文本长度: %d 字符\n", len(text))
		courses = utils.ParseCourseText(text)
	}

	fmt.Printf("🔍 [DEBUG] 解析到 %d 门课程\n", len(courses))

	if len(courses) == 0 {
		utils.BadRequest(c, "未能从HTML中识别到课程信息")
		return
	}

	// 🎯 只返回解析结果，不保存到数据库
	utils.Success(c, gin.H{
		"message": "解析成功（未保存）",
		"count":   len(courses),
		"courses": courses,
	})
}

// ParseCourseText 解析课程表文本
func ParseCourseText(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.Unauthorized(c, "用户未登录")
		return
	}

	var req struct {
		Text     string `json:"text" binding:"required"`
		Semester string `json:"semester" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误:  "+err.Error())
		return
	}

	// 解析文本
	courses := utils.ParseCourseText(req.Text)

	if len(courses) == 0 {
		utils.BadRequest(c, "未能从文本中识别到课程信息")
		return
	}

	db := database.GetDB()

	// 删除旧数据
	db.Where("user_id = ? AND semester = ? ", userID, req.Semester).Delete(&models.Course{})

	// 保存新数据
	for _, courseInfo := range courses {
		course := models.Course{
			UserID:       userID,
			CourseName:   courseInfo.CourseName,
			Teacher:      courseInfo.Teacher,
			Classroom:    courseInfo.Classroom,
			WeekDay:      courseInfo.WeekDay,
			StartWeek:    courseInfo.StartWeek,
			EndWeek:      courseInfo.EndWeek,
			StartSection: courseInfo.StartSection,
			EndSection:   courseInfo.EndSection,
			Semester:     req.Semester,
			WeekType:     courseInfo.WeekType,
		}

		if err := db.Create(&course).Error; err != nil {
			utils.InternalServerError(c, "保存课程失败", err)
			return
		}
	}

	utils.Success(c, gin.H{
		"message": "导入成功",
		"count":   len(courses),
		"courses": courses,
	})
}

// ManualAddCourse 手动添加单个课程
func ManualAddCourse(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.Unauthorized(c, "用户未登录")
		return
	}

	var req models.Course
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 设置用户ID
	req.UserID = userID

	// 验证必填字段
	if req.CourseName == "" || req.Semester == "" || req.WeekDay == 0 {
		utils.BadRequest(c, "课程名称、学期、星期不能为空")
		return
	}

	db := database.GetDB()
	if err := db.Create(&req).Error; err != nil {
		utils.InternalServerError(c, "保存课程失败", err)
		return
	}

	utils.Success(c, gin.H{
		"message": "添加成功",
		"course":  req,
	})
}

// GetCoursesByWeek 获取指定周的课程
func GetCoursesByWeek(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.Unauthorized(c, "用户未登录")
		return
	}

	semester := c.Query("semester")
	week := c.Query("week")

	if semester == "" || week == "" {
		utils.BadRequest(c, "学期和周次参数不能为空")
		return
	}

	db := database.GetDB()

	var courses []models.Course
	query := db.Where("user_id = ? AND semester = ? ", userID, semester)

	// 过滤周次
	query = query.Where("start_week <= ?  AND end_week >= ?", week, week)

	// 这里还需要根据 week_type 过滤单双周
	// 简化处理：先查出来再在应用层过滤
	if err := query.Order("week_day ASC, start_section ASC").Find(&courses).Error; err != nil {
		utils.InternalServerError(c, "查询失败", err)
		return
	}

	utils.Success(c, gin.H{
		"courses": courses,
	})
}

// GetAllCourses 获取所有课程
func GetAllCourses(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.Unauthorized(c, "用户未登录")
		return
	}

	semester := c.Query("semester")
	if semester == "" {
		utils.BadRequest(c, "学期参数不能为空")
		return
	}

	db := database.GetDB()

	var courses []models.Course
	if err := db.Where("user_id = ? AND semester = ?", userID, semester).
		Order("week_day ASC, start_section ASC").
		Find(&courses).Error; err != nil {
		utils.InternalServerError(c, "查询失败", err)
		return
	}

	utils.Success(c, gin.H{
		"courses": courses,
	})
}

// DeleteCourseSchedule 删除整个学期的课程表
func DeleteCourseSchedule(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.Unauthorized(c, "用户未登录")
		return
	}

	semester := c.Query("semester")
	if semester == "" {
		utils.BadRequest(c, "学期参数不能为空")
		return
	}

	db := database.GetDB()

	if err := db.Where("user_id = ? AND semester = ?", userID, semester).Delete(&models.Course{}).Error; err != nil {
		utils.InternalServerError(c, "删除失败", err)
		return
	}

	utils.Success(c, gin.H{
		"message": "删除成功",
	})
}

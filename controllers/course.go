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

	file, err := c.FormFile("html")
	if err != nil {
		utils.BadRequest(c, "请上传HTML文件")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".html" && ext != ".htm" {
		utils.BadRequest(c, "只支持HTML格式")
		return
	}

	tempDir := "temp"
	os.MkdirAll(tempDir, 0755)
	tempPath := filepath.Join(tempDir, fmt.Sprintf("%s_%d_%s", userID, time.Now().Unix(), file.Filename))

	err = c.SaveUploadedFile(file, tempPath)
	if err != nil {
		utils.InternalServerError(c, "保存文件失败", err)
		return
	}
	defer os.Remove(tempPath)

	courses, err := utils.ParseCourseHTML(tempPath)
	if err != nil {
		utils.InternalServerError(c, "解析HTML失败:  "+err.Error(), err)
		return
	}

	if len(courses) == 0 {
		utils.BadRequest(c, "未能从HTML中识别到课程信息")
		return
	}

	db := database.GetDB()
	db.Where("user_id = ?  AND semester = ?", userID, semester).Delete(&models.Course{})

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
	})
}

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

	req.UserID = userID

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
	query = query.Where("start_week <= ?  AND end_week >= ?", week, week)

	if err := query.Order("week_day ASC, start_section ASC").Find(&courses).Error; err != nil {
		utils.InternalServerError(c, "查询失败", err)
		return
	}

	utils.Success(c, gin.H{
		"courses": courses,
	})
}

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

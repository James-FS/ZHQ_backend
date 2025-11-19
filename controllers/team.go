package controllers

import (
	"zhq-backend/database"
	"zhq-backend/models"
	"zhq-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetTeamList 获取队伍列表
func GetTeamList(c *gin.Context) {
	var teams []models.Team
	var total int64

	//1.计算总数
	tx := database.GetDB().Model(&models.Team{})
	if err := tx.Count(&total).Error; err != nil {
		utils.InternalServerError(c, "获取队伍总数失败:", err)
		return
	}

	//2.获取列表（按创建时间排序）
	if err := tx.Order("created_at DESC").Find(&teams).Error; err != nil {
		utils.InternalServerError(c, "获取队伍列表失败:", err)
		return
	}

	//3.返回结果
	utils.Success(c, gin.H{
		"list":  teams,
		"total": total,
	})
}

// GetTeamDetails 获取队伍详情
func GetTeamDetails(c *gin.Context) {
	teamID := c.Param("team_id")
	if teamID == "" {
		utils.BadRequest(c, "teamID不能为空")
		return
	}

	var detail models.Team
	if err := database.DB.Where("team_id = ?", teamID).First(&detail).Error; err != nil {
		utils.BadRequest(c, "队伍不存在")
		return
	}
	utils.Success(c, detail)
}

// CreateTeam 创建队伍
func CreateTeam(c *gin.Context) {
	//1.获取当前登录用户ID(从认证中间件中获取)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}

	//2.绑定并校验请求参数
	var req struct {
		TeamName            string `json:"team_name" binding:"required,min=1,max=100"`
		Content             string `json:"content" binding:"required"`
		Pictures            string `json:"pictures"`
		MaxMembers          int    `json:"max_members" binding:"required,min=1,max=50"`
		Tags                string `json:"tags"`
		AnticipativeOutcome string `json:"anticipative_outcome"`
		RequireSkills       string `json:"require_skills"`
		RelativeContest     string `json:"relative_contest"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误:"+err.Error())
		return
	}

	// 3. 生成唯一TeamID（UUID）
	teamID := uuid.New().String()

	// 4.构造队伍数据
	team := models.Team{
		TeamID:              teamID,
		TeamName:            req.TeamName,
		Content:             req.Content,
		Pictures:            req.Pictures,
		CreatorID:           userID.(uint),
		MaxMembers:          req.MaxMembers,
		CurrentMembers:      1,
		Tags:                req.Tags,
		Status:              1, // 默认状态为招募中
		AnticipativeOutcome: req.AnticipativeOutcome,
		RequireSkills:       req.RequireSkills,
		RelativeContest:     req.RelativeContest,
	}

	//5.存入数据库
	if err := database.GetDB().Create(&team).Error; err != nil {
		utils.InternalServerError(c, "创建队伍失败:", err)
		return
	}

	//6.返回结果
	utils.SuccessWithMessage(c, "队伍创建成功", team)
}

func UpdateTeam(c *gin.Context) {
	teamID := c.Param("team_id")
	if teamID == "" {
		utils.BadRequest(c, "teamID不能为空")
		return
	}

	var team models.Team
	if err := database.DB.Where("team_id = ?", teamID).First(&team).Error; err != nil {
		utils.BadRequest(c, "队伍不存在")
		return
	}

	if err := c.ShouldBindJSON(&team); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := database.DB.Save(&team).Error; err != nil {
		utils.InternalServerError(c, "更新队伍失败: ", err)
		return
	}
	utils.SuccessWithMessage(c, "编辑成功", team)
}

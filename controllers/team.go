package controllers

import (
	"zhq-backend/database"
	"zhq-backend/models"
	"zhq-backend/utils"

	"github.com/gin-gonic/gin"
)

// GetTeamList 获取队伍列表
func GetTeamList(c *gin.Context) {
	var teams []models.Team
	var total int64

	//1.计算总数
	tx := database.GetDB().Model(&models.Team{})
	if err := tx.Count(&total).Error; err != nil {
		utils.InternalServerError(c, "获取队伍总数失败:"+err.Error())
		return
	}

	//2.获取列表（按创建时间排序）
	if err := tx.Order("created_at DESC").Find(&teams).Error; err != nil {
		utils.InternalServerError(c, "获取队伍列表失败:"+err.Error())
		return
	}

	//3.返回结果
	utils.Success(c, gin.H{
		"list":  teams,
		"total": total,
	})
}

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
		utils.InternalServerError(c, "更新队伍失败: "+err.Error())
		return
	}
	utils.SuccessWithMessage(c, "编辑成功", team)
}

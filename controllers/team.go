package controllers

import (
	"strconv"
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
	teamID := c.Query("id")
	if teamID == "" {
		utils.BadRequest(c, "teamID不能为空")
		return
	}
	id, err := strconv.ParseUint(teamID, 10, 64)
	if err != nil {
		utils.BadRequest(c, "teamID格式错误，必须是整数")
		return
	}
	var detail models.Team
	if err := database.DB.First(&detail, uint(id)).Error; err != nil {
		utils.BadRequest(c, "队伍不存在")
		return
	}
	utils.Success(c, detail)
}

package controllers

import (
	"strconv"
	"zhq-backend/database"
	"zhq-backend/models"
	"zhq-backend/utils"

	"github.com/gin-gonic/gin"
)

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
	var detail models.TeamDetail
	if err := database.DB.First(&detail, uint(id)).Error; err != nil {
		utils.BadRequest(c, "队伍不存在")
		return
	}
	utils.Success(c, detail)
}

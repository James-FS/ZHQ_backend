package controllers

import (
	"zhq-backend/database"
	"zhq-backend/models"
	"zhq-backend/utils"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetTeamList 获取队伍列表
func GetTeamList(c *gin.Context) {
	var teams []models.Team
	var total int64

	//1.获取查询参数(可选)
	teamID := c.Query("team_id")
	teamName := c.Query("team_name")
	tags := c.Query("tags")
	CreatorID := c.Query("creator_id")
	creatorNickName := c.Query("creator_nickname")
	content := c.Query("content")

	//2.初始化查询
	tx := database.GetDB().Model(&models.Team{})

	//3.应用过滤条件
	if teamID != "" {
		tx = tx.Where("team_id = ?", teamID) //精确查询
	}
	if teamName != "" {
		tx = tx.Where("team_name LIKE ?", "%"+teamName+"%") //模糊查
	}
	if tags != "" {
		tx = tx.Where("JSON_CONTAINS(tags, ?, '$')", "'"+tags+"'") //模糊查
	}
	if CreatorID != "" {
		tx = tx.Where("creator_id = ?", CreatorID) //精确查询
	}
	if content != "" {
		tx = tx.Where("content LIKE ?", "%"+content+"%") //模糊查询
	}

	// 通过关联查询用户表获取CreatorID列表
	if creatorNickName != "" {
		var userIDs []string

		// 查询匹配昵称的用户ID列表
		if err := database.GetDB().
			Model(&models.User{}). // 显式指定查询的模型
			Where("nickname LIKE ?", "%"+creatorNickName+"%").
			Pluck("user_id", &userIDs).Error; err != nil {
			utils.InternalServerError(c, "获取创建者信息失败:", err)
			return
		}

		if len(userIDs) > 0 {
			tx = tx.Where("creator_id IN (?)", userIDs)
		} else {
			// 如果没有匹配的用户，直接返回空结果
			utils.Success(c, gin.H{
				"list":  []models.Team{},
				"total": 0,
			})
			return
		}
	}

	// 5.计算总数
	if err := tx.Count(&total).Error; err != nil {
		utils.InternalServerError(c, "获取队伍总数失败:", err)
		return
	}

	// 6.获取列表（按创建时间排序）
	if err := tx.Order("created_at DESC").Find(&teams).Error; err != nil {
		utils.InternalServerError(c, "获取队伍列表失败:", err)
		return
	}

	// 7.获取队伍创建者UserID
	var creatorIDs []string
	for _, team := range teams {
		creatorIDs = append(creatorIDs, team.CreatorID)
	}

	// 8.批量查询用户信息（根据string类型的UserID查询）
	var users []models.User
	if err := database.GetDB().Where("user_id IN (?)", creatorIDs).Find(&users).Error; err != nil {
		utils.InternalServerError(c, "获取创建者信息失败:", err)
		return
	}

	// 9.将用户信息映射为map，key为string类型的UserID
	userMap := make(map[string]models.User)
	for _, user := range users {
		userMap[user.UserID] = user // 假设User模型中用户ID字段是UserID(string类型)
	}

	// 10.组装包含用户信息的响应数据
	type TeamWithCreator struct {
		models.Team
		CreatorNickname string   `json:"creator_nickname"`
		CreatorAvatar   string   `json:"creator_avatar"`
		TagsArray       []string `json:"tags"`
	}

	var resultList []TeamWithCreator
	for _, team := range teams {
		creator, exists := userMap[team.CreatorID]
		creatorNickname := ""
		creatorAvatar := ""
		if exists {
			creatorNickname = creator.Nickname
			creatorAvatar = creator.Avatar
		}

		var tagsArray []string
		if team.Tags != "" {
			err := json.Unmarshal([]byte(team.Tags), &tagsArray)
			if err != nil {
				return
			}
		}

		resultList = append(resultList, TeamWithCreator{
			Team:            team,
			CreatorNickname: creatorNickname,
			CreatorAvatar:   creatorAvatar,
			TagsArray:       tagsArray,
		})
	}

	// 11.返回结果
	utils.Success(c, gin.H{
		"list":  resultList,
		"total": total,
	})
}

// GetTeamDetails 获取队伍详情
func GetTeamDetails(c *gin.Context) {
	teamID := c.Query("team_id")
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
		TeamName            string   `json:"team_name" binding:"required,min=1,max=100"`
		Content             string   `json:"content" binding:"required"`
		Pictures            string   `json:"pictures"`
		MaxMembers          int      `json:"max_members" binding:"required,min=1,max=50"`
		Tags                []string `json:"tags"`
		AnticipativeOutcome string   `json:"anticipative_outcome"`
		RequireSkills       string   `json:"require_skills"`
		RelativeContest     string   `json:"relative_contest"`
		ProjectCycle        string   `json:"project_cycle" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误:"+err.Error())
		return
	}

	// 3. 生成唯一TeamID（UUID）
	teamID := uuid.New().String()

	// 新增：将 tags 切片转换为 JSON 字符串
	tagsJSON, err := json.Marshal(req.Tags)
	if err != nil {
		utils.BadRequest(c, "标签格式错误: "+err.Error())
		return
	}

	// 4.构造队伍数据
	team := models.Team{
		TeamID:              teamID,
		TeamName:            req.TeamName,
		Content:             req.Content,
		Pictures:            req.Pictures,
		CreatorID:           userID.(string),
		MaxMembers:          req.MaxMembers,
		CurrentMembers:      1,
		Tags:                string(tagsJSON),
		Status:              1, // 默认状态为招募中
		AnticipativeOutcome: req.AnticipativeOutcome,
		RequireSkills:       req.RequireSkills,
		RelativeContest:     req.RelativeContest,
		ProjectCycle:        req.ProjectCycle,
	}

	//5.存入数据库
	if err := database.GetDB().Create(&team).Error; err != nil {
		utils.InternalServerError(c, "创建队伍失败:", err)
		return
	}

	//6.返回结果
	utils.SuccessWithMessage(c, "队伍创建成功", team)
}

// UpdateTeam 编辑队伍信息
func UpdateTeam(c *gin.Context) {
	teamID := c.Query("id")
	if teamID == "" {
		utils.BadRequest(c, "teamID不能为空")
		return
	}

	var team models.Team
	if err := database.DB.Where("team_id = ?", teamID).First(&team).Error; err != nil {
		utils.BadRequest(c, "队伍不存在")
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	allowedFields := map[string]bool{
		"team_name":     true,
		"description":   true,
		"category":      true,
		"max_members":   true,
		"status":        true,
		"project_cycle": true,
	}

	// ③ 过滤掉不允许更新的字段
	for key := range updateData {
		if !allowedFields[key] {
			delete(updateData, key)
		}
	}

	// ④ 只更新指定的字段
	if err := database.DB.Model(&team).Updates(updateData).Error; err != nil {
		utils.InternalServerError(c, "更新队伍失败: ", err)
		return
	}

	utils.SuccessWithMessage(c, "编辑成功", team)
}

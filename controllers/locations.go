package controllers

import (
	"fmt"
	"zhq-backend/database"
	"zhq-backend/models"
	"zhq-backend/utils"

	"github.com/gin-gonic/gin"
)

// GetLocations 获取所有校园地点（支持分类筛选）
// GET /api/v1/locations?category=teaching&page=1&limit=100
func GetLocations(c *gin.Context) {
	category := c.Query("category")
	page := 1
	limit := 100

	// 解析分页参数
	if p := c.Query("page"); p != "" {
		var pageNum int
		if _, err := fmt.Sscanf(p, "%d", &pageNum); err == nil && pageNum > 0 {
			page = pageNum
		}
	}
	if l := c.Query("limit"); l != "" {
		var limitNum int
		if _, err := fmt.Sscanf(l, "%d", &limitNum); err == nil && limitNum > 0 && limitNum <= 100 {
			limit = limitNum
		}
	}

	var locations []models.Location
	query := database.DB.Where("status = ?", 1)

	// 如果指定了分类，则过滤
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 获取总数
	var total int64
	query.Model(&models.Location{}).Count(&total)

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&locations).Error; err != nil {
		utils.InternalServerError(c, "获取地点列表失败", err)
		return
	}

	// 转换为响应格式
	var response []models.LocationListResponse
	for _, loc := range locations {
		response = append(response, models.LocationListResponse{
			ID:        loc.ID,
			Name:      loc.Name,
			Category:  loc.Category,
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
		})
	}

	c.JSON(200, gin.H{
		"code": 0,
		"data": response,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetLocationsByCategory 按分类获取地点
// GET /api/v1/locations/category/:category
func GetLocationsByCategory(c *gin.Context) {
	category := c.Param("category")

	var locations []models.Location
	if err := database.DB.Where("category = ? AND status = ?", category, 1).Find(&locations).Error; err != nil {
		utils.InternalServerError(c, "获取地点列表失败", err)
		return
	}
	var response []models.LocationListResponse
	for _, loc := range locations {
		response = append(response, models.LocationListResponse{
			ID:        loc.ID,
			Name:      loc.Name,
			Category:  loc.Category,
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
		})
	}

	utils.Success(c, response)
}

// GetLocationDetail 获取地点详情
// GET /api/v1/locations/:id
func GetLocationDetail(c *gin.Context) {
	id := c.Param("id")

	var location models.Location
	if err := database.DB.Where("id = ? AND status = ?", id, 1).First(&location).Error; err != nil {
		utils.NotFound(c, "地点不存在")
		return
	}

	response := models.LocationDetailResponse{
		ID:        location.ID,
		Name:      location.Name,
		Category:  location.Category,
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
		Tags:      location.Tags,
	}

	utils.Success(c, response)
}

// SearchLocations 搜索地点
// GET /api/v1/locations/search?keyword=图书馆
func SearchLocations(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		utils.BadRequest(c, "搜索关键词不能为空")
		return
	}

	var locations []models.Location
	query := database.DB.Where("status = ? AND name LIKE ?",
		1, "%"+keyword+"%")

	if err := query.Find(&locations).Error; err != nil {
		utils.InternalServerError(c, "搜索地点失败", err)
		return
	}

	var response []models.LocationListResponse
	for _, loc := range locations {
		response = append(response, models.LocationListResponse{
			ID:        loc.ID,
			Name:      loc.Name,
			Category:  loc.Category,
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
		})
	}

	utils.Success(c, response)
}

// CreateLocation 创建地点（管理员）
func CreateLocation(c *gin.Context) {
	var location models.Location
	if err := c.ShouldBindJSON(&location); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 验证必填字段
	if location.Name == "" || location.Category == "" {
		utils.BadRequest(c, "地点名称和分类不能为空")
		return
	}

	location.Status = 1
	if err := database.DB.Create(&location).Error; err != nil {
		utils.InternalServerError(c, "创建地点失败", err)
		return
	}

	utils.Success(c, gin.H{
		"id":      location.ID,
		"message": "地点创建成功",
	})
}

// UpdateLocation 更新地点信息（管理员）
func UpdateLocation(c *gin.Context) {
	id := c.Param("id")

	var location models.Location
	if err := database.DB.Where("id = ?", id).First(&location).Error; err != nil {
		utils.NotFound(c, "地点不存在")
		return
	}

	if err := c.ShouldBindJSON(&location); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := database.DB.Model(&location).Updates(location).Error; err != nil {
		utils.InternalServerError(c, "更新地点失败", err)
		return
	}

	utils.Success(c, gin.H{
		"message": "地点更新成功",
	})
}

// DeleteLocation 删除地点（管理员）
func DeleteLocation(c *gin.Context) {
	id := c.Param("id")

	var location models.Location
	if err := database.DB.Where("id = ?", id).First(&location).Error; err != nil {
		utils.NotFound(c, "地点不存在")
		return
	}

	// 软删除（仅标记为无效）
	if err := database.DB.Model(&location).Update("status", 0).Error; err != nil {
		utils.InternalServerError(c, "删除地点失败", err)
		return
	}

	utils.Success(c, gin.H{
		"message": "地点删除成功",
	})
}

// GetLocationCategories 获取所有地点分类及数量
// GET /api/v1/locations/categories
func GetLocationCategories(c *gin.Context) {
	type CategoryCount struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}

	var categories []CategoryCount
	if err := database.DB.Model(&models.Location{}).
		Where("status = ?", 1).
		Select("category, COUNT(*) as count").
		Group("category").
		Scan(&categories).Error; err != nil {
		utils.InternalServerError(c, "获取分类失败", err)
		return
	}

	utils.Success(c, categories)
}

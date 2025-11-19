package routes

import (
	"zhq-backend/controllers"
	"zhq-backend/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// 添加CORS中间件
	r.Use(middleware.CORS())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "ZHQ Backend Server is running",
			"time":    "2025-11-17 01:55:15",
		})
	})

	// API版本1
	v1 := r.Group("/api/v1")
	{
		// 认证相关路由（无需登录）
		auth := v1.Group("/auth")
		{
			auth.POST("/login", controllers.WeChatLogin)
		}

		//广场页面的队伍列表（无需登录）
		v1.GET("/teams", controllers.GetTeamList)
		v1.GET("/team/details/:team_id", controllers.GetTeamDetails)

		// 需要认证的路由
		authorized := v1.Group("/")
		authorized.Use(middleware.AuthRequired())
		{
			// 用户相关
			user := authorized.Group("/user")
			{
				user.GET("/profile", controllers.GetUserProfile)
				user.PUT("/profile", controllers.UpdateUserProfile)
			}
			// 组队相关
			team := authorized.Group("/team/edit")
			{
				team.PUT("/:team_id", controllers.UpdateTeam)

			}
		}
	}
}

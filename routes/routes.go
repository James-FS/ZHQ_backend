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

			// 队伍相关
			teams := authorized.Group("/teams")
			{
				teams.POST("", controllers.CreateTeam) // 创建队伍
				// 后续可添加：修改队伍、解散队伍、申请加入等接口
			}
		}
	}
}

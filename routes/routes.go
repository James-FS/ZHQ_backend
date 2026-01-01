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
			auth.POST("/phone-login", controllers.PhonePasswordLogin)
			auth.POST("/register", controllers.RegisterByPhone)
		}

		//广场页面的队伍列表（无需登录）
		v1.GET("/teams", controllers.GetTeamList)
		v1.GET("/team/details/:team_id", controllers.GetTeamDetails)

		// 路线规划（无需登录）
		v1.POST("/route", controllers.GetRoute)

		// 需要认证的路由
		authorized := v1.Group("/")
		authorized.Use(middleware.AuthRequired())
		{
			// 用户相关
			user := authorized.Group("/user")
			{
				user.GET("", controllers.GetUserProfile)
				user.PUT("/profile", controllers.UpdateUserProfile)
				user.GET("/collection", controllers.GetUserCollection)
				user.GET("/collection/status", controllers.CheckUserCollection)
				user.POST("/collection", controllers.AddUserCollection)
				user.DELETE("/collection", controllers.RemoveUserCollection)
				user.PUT("/uploadAvatar", controllers.UploadAvatar)
				user.POST("/uploadResume", controllers.UploadResume)
			}

			// 队伍相关
			teams := authorized.Group("/teams")
			{
				teams.POST("", controllers.CreateTeam)              // 创建队伍
				teams.PUT("/edit/:team_id", controllers.UpdateTeam) //编辑队伍
				teams.GET("/details", controllers.GetTeamDetails)
				teams.POST("/:team_id/join", controllers.JoinTeam) // 加入队伍
				teams.GET("/my-teams", controllers.GetMyTeams)
				// 后续可添加：修改队伍、解散队伍、申请加入等接口
			}

			// 消息相关
			chat := authorized.Group("/chat")
			{
				chat.GET("/ws", controllers.WebSocketHandler)
				chat.GET("/online", controllers.GetOnlineUsers)
				chat.GET("/check-online", controllers.CheckUserOnline)
				chat.GET("/history", controllers.GetChatHistory)
				chat.GET("/list", controllers.GetChatList)
			}

			// 好友相关
			friends := authorized.Group("/friend")
			{
				friends.GET("", controllers.GetFriendList)              //获取好友列表
				friends.POST("", controllers.AddFriend)                 //发送好友请求
				friends.GET("/requests", controllers.GetFriendRequests) //获取好友请求
				friends.POST("/accept", controllers.AcceptFriend)       //接受好友请求
				friends.POST("/reject", controllers.RejectFriend)       //拒绝好友请求
				friends.GET("/check", controllers.CheckFriendship)      // 检查是否为好友

			}

			// 课程表
			course := authorized.Group("/course")
			{
				course.POST("/upload-pdf", controllers.UploadCoursePDF)
				course.POST("/upload-html", controllers.UploadCourseHTML) // ← 添加这行
				course.POST("/parse-text", controllers.ParseCourseText)
				course.POST("/manual-add", controllers.ManualAddCourse)
				course.GET("/week", controllers.GetCoursesByWeek)
				course.GET("/all", controllers.GetAllCourses)
				course.DELETE("/schedule", controllers.DeleteCourseSchedule)
				course.POST("/test-parse-html", controllers.TestParseHTML)
			}
		}
	}
}

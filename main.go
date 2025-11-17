package main

import (
	"log"
	"zhq-backend/config"
	"zhq-backend/database"
	"zhq-backend/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// 初始化配置
	config.Init()

	// 连接数据库
	database.Init()

	// 创建Gin引擎
	r := gin.Default()

	// 设置路由
	routes.SetupRoutes(r)

	// 启动服务器
	port := config.GetString("server.port")
	log.Printf("Server starting on port %s", port)
	log.Printf("Server will be available at: http://localhost:%s", port)
	r.Run(":" + port)
}

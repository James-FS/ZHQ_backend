package database

import (
	"fmt"
	"log"
	"zhq-backend/config"
	"zhq-backend/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error

	// 构建数据库连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.GetString("database.user"),
		config.GetString("database.password"),
		config.GetString("database.host"),
		config.GetString("database.port"),
		config.GetString("database.name"),
	)

	// 连接数据库
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		log.Printf("Please ensure MySQL is running and database 'zhq' exists")
		log.Fatal("Database connection failed")
	}

	log.Println("Database connected successfully to 'zhq'")

	// 自动迁移数据库表
	autoMigrate()
}

func autoMigrate() {
	err := DB.AutoMigrate(
		&models.User{},
		// 添加其他模型...
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migration completed for 'zhq' database")
}

// 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

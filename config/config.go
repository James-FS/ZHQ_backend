package config

import (
	"os"

	"github.com/spf13/viper"
)

func Init() {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")

	// 从环境变量读取配置
	viper.AutomaticEnv()

	// 数据库配置
	viper.Set("database.host", getEnv("DB_HOST", "localhost"))
	viper.Set("database.port", getEnv("DB_PORT", "3306"))
	viper.Set("database.user", getEnv("DB_USER", "root"))
	viper.Set("database.password", getEnv("DB_PASSWORD", ""))
	viper.Set("database.name", getEnv("DB_NAME", "zhq")) // 修改为 zhq

	// 服务器配置
	viper.Set("server.port", getEnv("SERVER_PORT", "8080"))
	viper.Set("server.host", getEnv("SERVER_HOST", "0.0.0.0"))

	// JWT配置
	viper.Set("jwt.secret", getEnv("JWT_SECRET", "default_secret"))

	// 微信配置
	viper.Set("wechat.appid", getEnv("WECHAT_APPID", ""))
	viper.Set("wechat.secret", getEnv("WECHAT_SECRET", ""))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetString(key string) string {
	return viper.GetString(key)
}

func GetInt(key string) int {
	return viper.GetInt(key)
}

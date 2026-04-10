package config

import (
	"os"
	"strconv"
)

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// GetMySQLConfig 获取 MySQL 配置（从环境变量）
func GetMySQLConfig() MySQLConfig {
	port, _ := strconv.Atoi(getEnv("MYSQL_PORT", "3306"))

	return MySQLConfig{
		Host:     getEnv("MYSQL_HOST", "localhost"),
		Port:     port,
		User:     getEnv("MYSQL_USER", "root"),
		Password: getEnv("MYSQL_PASSWORD", ""),
		Database: getEnv("MYSQL_DATABASE", "idle_rpg"),
	}
}

// GetRedisConfig 获取 Redis 配置（从环境变量）
func GetRedisConfig() RedisConfig {
	db, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return RedisConfig{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       db,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

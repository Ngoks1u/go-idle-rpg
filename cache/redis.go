package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	ctx    = context.Background()
)

// Config Redis 配置
type Config struct {
	Addr     string
	Password string
	DB       int
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
}

// Init 初始化 Redis 连接
func Init(cfg Config) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	_, err := Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Println("Redis connected successfully")
	return nil
}

// Close 关闭连接
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// PlayerKey 生成玩家缓存 key
func PlayerKey(playerID int64) string {
	return fmt.Sprintf("player:%d", playerID)
}

// EquipmentKey 生成装备列表缓存 key
func EquipmentKey(playerID int64) string {
	return fmt.Sprintf("equipment:%d", playerID)
}

// GetPlayerJSON 获取缓存的玩家 JSON
func GetPlayerJSON(playerID int64) (string, error) {
	return Client.Get(ctx, PlayerKey(playerID)).Result()
}

// SetPlayerJSON 设置玩家缓存
func SetPlayerJSON(playerID int64, data string, ttl time.Duration) error {
	return Client.Set(ctx, PlayerKey(playerID), data, ttl).Err()
}

// DeletePlayer 删除玩家缓存
func DeletePlayer(playerID int64) error {
	return Client.Del(ctx, PlayerKey(playerID)).Err()
}

// GetEquipmentJSON 获取缓存的装备 JSON
func GetEquipmentJSON(playerID int64) (string, error) {
	return Client.Get(ctx, EquipmentKey(playerID)).Result()
}

// SetEquipmentJSON 设置装备缓存
func SetEquipmentJSON(playerID int64, data string, ttl time.Duration) error {
	return Client.Set(ctx, EquipmentKey(playerID), data, ttl).Err()
}

// DeleteEquipment 删除装备缓存
func DeleteEquipment(playerID int64) error {
	return Client.Del(ctx, EquipmentKey(playerID)).Err()
}

// IncrBy 原子递增
func IncrBy(key string, value int64) (int64, error) {
	return Client.IncrBy(ctx, key, value).Result()
}

// GetInt 获取整数值
func GetInt(key string) (int64, error) {
	return Client.Get(ctx, key).Int64()
}

// SetExpire 设置过期时间
func SetExpire(key string, ttl time.Duration) error {
	return Client.Expire(ctx, key, ttl).Err()
}

// Exists 检查 key 是否存在
func Exists(key string) (bool, error) {
	count, err := Client.Exists(ctx, key).Result()
	return count > 0, err
}

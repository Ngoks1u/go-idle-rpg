package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go-idle-rpg/cache"
	"go-idle-rpg/models"
)

var DB *sql.DB

// 缓存 TTL
const (
	PlayerCacheTTL    = 5 * time.Minute
	EquipmentCacheTTL = 10 * time.Minute
)

// Config 数据库配置
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "",
		Database: "idle_rpg",
	}
}

// Init 初始化 MySQL 数据库
func Init(cfg Config) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// 测试连接
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// 创建表
	err = createTables()
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("MySQL database initialized successfully")
	return nil
}

// Close 关闭数据库连接
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// createTables 创建所有表
func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS players (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(50) UNIQUE NOT NULL,
			level INT DEFAULT 1,
			exp BIGINT DEFAULT 0,
			health INT DEFAULT 100,
			max_health INT DEFAULT 100,
			attack INT DEFAULT 10,
			defense INT DEFAULT 2,
			gold BIGINT DEFAULT 0,
			is_fighting BOOLEAN DEFAULT FALSE,
			auto_fight BOOLEAN DEFAULT FALSE,
			last_fight TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_name (name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS equipment (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			type INT NOT NULL,
			quality INT NOT NULL,
			level INT NOT NULL,
			attack INT DEFAULT 0,
			defense INT DEFAULT 0,
			player_id BIGINT NOT NULL,
			equipped BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_player_id (player_id),
			INDEX idx_equipped (equipped),
			FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	for _, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// CreatePlayer 创建新玩家
func CreatePlayer(name string) (*models.Player, error) {
	query := `INSERT INTO players (name, level, health, max_health, attack, defense)
	          VALUES (?, 1, 100, 100, 10, 2)`

	result, err := DB.Exec(query, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	id, _ := result.LastInsertId()
	return GetPlayer(id)
}

// GetPlayer 获取玩家 by ID
func GetPlayer(id int64) (*models.Player, error) {
	// 先尝试从缓存获取
	if cache.Client != nil {
		cachedData, err := cache.GetPlayerJSON(id)
		if err == nil && cachedData != "" {
			player := &models.Player{}
			if err := json.Unmarshal([]byte(cachedData), player); err == nil {
				return player, nil
			}
		}
	}

	query := `SELECT id, name, level, exp, health, max_health, attack, defense,
	                 gold, is_fighting, auto_fight, last_fight, created_at, updated_at
	          FROM players WHERE id = ?`

	row := DB.QueryRow(query, id)

	player := &models.Player{}
	err := row.Scan(
		&player.ID, &player.Name, &player.Level, &player.Exp,
		&player.Health, &player.MaxHealth, &player.Attack, &player.Defense,
		&player.Gold, &player.IsFighting, &player.AutoFight, &player.LastFight,
		&player.CreatedAt, &player.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	// 写入缓存
	if cache.Client != nil {
		if data, err := json.Marshal(player); err == nil {
			cache.SetPlayerJSON(id, string(data), PlayerCacheTTL)
		}
	}

	return player, nil
}

// GetPlayerByName 根据名字获取玩家
func GetPlayerByName(name string) (*models.Player, error) {
	query := `SELECT id, name, level, exp, health, max_health, attack, defense,
	                 gold, is_fighting, auto_fight, last_fight, created_at, updated_at
	          FROM players WHERE name = ?`

	row := DB.QueryRow(query, name)

	player := &models.Player{}
	err := row.Scan(
		&player.ID, &player.Name, &player.Level, &player.Exp,
		&player.Health, &player.MaxHealth, &player.Attack, &player.Defense,
		&player.Gold, &player.IsFighting, &player.AutoFight, &player.LastFight,
		&player.CreatedAt, &player.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	return player, nil
}

// UpdatePlayer 更新玩家
func UpdatePlayer(player *models.Player) error {
	query := `UPDATE players SET
	          level = ?, exp = ?, health = ?, max_health = ?,
	          attack = ?, defense = ?, gold = ?, is_fighting = ?,
	          auto_fight = ?, last_fight = ?
	          WHERE id = ?`

	_, err := DB.Exec(query, player.Level, player.Exp, player.Health, player.MaxHealth,
		player.Attack, player.Defense, player.Gold, player.IsFighting,
		player.AutoFight, player.LastFight, player.ID)

	if err != nil {
		return fmt.Errorf("failed to update player: %w", err)
	}

	// 更新缓存
	if cache.Client != nil {
		if data, err := json.Marshal(player); err == nil {
			cache.SetPlayerJSON(player.ID, string(data), PlayerCacheTTL)
		}
	}

	return nil
}

// SaveEquipment 保存装备
func SaveEquipment(equip *models.Equipment, playerID int64) error {
	query := `INSERT INTO equipment (name, type, quality, level, attack, defense, player_id, equipped)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := DB.Exec(query, equip.Name, equip.Type, equip.Quality, equip.Level,
		equip.Attack, equip.Defense, playerID, equip.Equipped)
	if err != nil {
		return fmt.Errorf("failed to save equipment: %w", err)
	}

	id, _ := result.LastInsertId()
	equip.ID = id

	// 清除装备缓存
	if cache.Client != nil {
		cache.DeleteEquipment(playerID)
	}

	return nil
}

// GetPlayerEquipment 获取玩家所有装备
func GetPlayerEquipment(playerID int64) ([]*models.Equipment, error) {
	// 先尝试从缓存获取
	if cache.Client != nil {
		cachedData, err := cache.GetEquipmentJSON(playerID)
		if err == nil && cachedData != "" {
			var equipment []*models.Equipment
			if err := json.Unmarshal([]byte(cachedData), &equipment); err == nil {
				return equipment, nil
			}
		}
	}

	query := `SELECT id, name, type, quality, level, attack, defense, equipped
	          FROM equipment WHERE player_id = ? ORDER BY quality DESC, level DESC`

	rows, err := DB.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}
	defer rows.Close()

	var equipment []*models.Equipment
	for rows.Next() {
		equip := &models.Equipment{}
		err := rows.Scan(&equip.ID, &equip.Name, &equip.Type, &equip.Quality,
			&equip.Level, &equip.Attack, &equip.Defense, &equip.Equipped)
		if err != nil {
			log.Printf("Error scanning equipment: %v", err)
			continue
		}
		equipment = append(equipment, equip)
	}

	// 写入缓存
	if cache.Client != nil && len(equipment) > 0 {
		if data, err := json.Marshal(equipment); err == nil {
			cache.SetEquipmentJSON(playerID, string(data), EquipmentCacheTTL)
		}
	}

	return equipment, nil
}

// DeletePlayer 删除玩家
func DeletePlayer(id int64) error {
	_, err := DB.Exec("DELETE FROM players WHERE id = ?", id)

	// 清除缓存
	if cache.Client != nil {
		cache.DeletePlayer(id)
	}

	return err
}

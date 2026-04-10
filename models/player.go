package models

import (
	"database/sql"
	"time"
)

// Player 玩家
type Player struct {
	ID          int64          `db:"id"`
	Name        string         `db:"name"`
	Level       int            `db:"level"`
	Exp         int64          `db:"exp"`
	Health      int            `db:"health"`
	MaxHealth   int            `db:"max_health"`
	Attack      int            `db:"attack"`
	Defense     int            `db:"defense"`
	Gold        int64          `db:"gold"`
	IsFighting  bool           `db:"is_fighting"`
	AutoFight   bool           `db:"auto_fight"`
	LastFight   sql.NullTime   `db:"last_fight"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

// ExpToLevel 计需要的经验值
func ExpToLevel(level int) int64 {
	return int64(level * level * 100)
}

// CanLevelUp 检查是否可以升级
func (p *Player) CanLevelUp() bool {
	nextLevelExp := ExpToLevel(p.Level + 1)
	return p.Exp >= nextLevelExp
}

// LevelUp 升级
func (p *Player) LevelUp() {
	if !p.CanLevelUp() {
		return
	}
	p.Level++
	p.MaxHealth = 100 + p.Level*20
	p.Health = p.MaxHealth
	p.Attack = 10 + p.Level*5
	p.Defense = 2 + p.Level*2
}

// AddExp 添加经验
func (p *Player) AddExp(exp int64) {
	p.Exp += exp
	for p.CanLevelUp() {
		p.LevelUp()
	}
}

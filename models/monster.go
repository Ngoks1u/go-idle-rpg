package models

import (
	"math/rand"
	"time"
)

// MonsterType 怪物类型
type MonsterType int

const (
	MonsterSlime   MonsterType = iota // 史莱姆
	MonsterGoblin                      // 哥布林
	MonsterOrc                         // 兽人
	MonsterDragon                      // 龙
)

// Monster 怪物定义
type Monster struct {
	ID       int64     `db:"id"`
	Name     string    `db:"name"`
	Type     MonsterType `db:"type"`
	Level    int       `db:"level"`
	Health   int       `db:"health"`
	MaxHealth int      `db:"max_health"`
	Attack   int       `db:"attack"`
	Defense  int       `db:"defense"`
	ExpReward int64    `db:"exp_reward"`
	GoldReward int64   `db:"gold_reward"`
	DropRate float64   `db:"drop_rate"`
}

// GenerateMonster 生成适合玩家等级的怪物
func GenerateMonster(playerLevel int) *Monster {
	rand.Seed(time.Now().UnixNano())

	monsterTypes := []MonsterType{
		MonsterSlime, MonsterSlime, MonsterSlime,
		MonsterGoblin, MonsterGoblin,
		MonsterOrc,
	}

	if playerLevel >= 10 {
		monsterTypes = append(monsterTypes, MonsterDragon)
	}

	mType := monsterTypes[rand.Intn(len(monsterTypes))]
	levelDiff := rand.Intn(3) - 1
	monsterLevel := playerLevel + levelDiff
	if monsterLevel < 1 {
		monsterLevel = 1
	}

	var name string
	var baseHealth, baseAttack, baseDefense int
	var expReward, goldReward int64
	var dropRate float64

	switch mType {
	case MonsterSlime:
		name = "史莱姆"
		baseHealth, baseAttack, baseDefense = 30, 5, 1
		expReward, goldReward = 10, 5
		dropRate = 0.1
	case MonsterGoblin:
		name = "哥布林"
		baseHealth, baseAttack, baseDefense = 50, 10, 3
		expReward, goldReward = 20, 15
		dropRate = 0.2
	case MonsterOrc:
		name = "兽人"
		baseHealth, baseAttack, baseDefense = 80, 18, 5
		expReward, goldReward = 50, 30
		dropRate = 0.3
	case MonsterDragon:
		name = "龙"
		baseHealth, baseAttack, baseDefense = 200, 40, 15
		expReward, goldReward = 200, 100
		dropRate = 0.5
	}

	levelMultiplier := float64(monsterLevel)

	return &Monster{
		Name:       name,
		Type:       mType,
		Level:      monsterLevel,
		Health:     int(float64(baseHealth) * levelMultiplier),
		MaxHealth:  int(float64(baseHealth) * levelMultiplier),
		Attack:     int(float64(baseAttack) * levelMultiplier),
		Defense:    int(float64(baseDefense) * levelMultiplier),
		ExpReward:  int64(float64(expReward) * levelMultiplier),
		GoldReward: int64(float64(goldReward) * levelMultiplier),
		DropRate:   dropRate,
	}
}

// Damage 计算受到的伤害
func (m *Monster) Damage(attack int) int {
	damage := attack - m.Defense
	if damage < 1 {
		damage = 1
	}
	return damage
}

// IsAlive 检查怪物是否存活
func (m *Monster) IsAlive() bool {
	return m.Health > 0
}

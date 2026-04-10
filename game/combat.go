package game

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go-idle-rpg/db"
	"go-idle-rpg/models"
)

// Game 游戏实例
type Game struct {
	Player      *models.Player
	Equipment   []*models.Equipment
	AutoTicker  *time.Ticker
	AutoStopped chan bool
}

// NewGame 创建新游戏
func NewGame(playerName string) (*Game, error) {
	// 检查玩家是否存在
	player, err := db.GetPlayerByName(playerName)
	if err != nil {
		// 创建新玩家
		player, err = db.CreatePlayer(playerName)
		if err != nil {
			return nil, fmt.Errorf("failed to create player: %w", err)
		}
	}

	// 加载装备
	equipment, err := db.GetPlayerEquipment(player.ID)
	if err != nil {
		log.Printf("Warning: failed to load equipment: %v", err)
	}

	g := &Game{
		Player:      player,
		Equipment:   equipment,
		AutoStopped: make(chan bool),
	}

	return g, nil
}

// ApplyEquipmentBonuses 应用装备属性加成
func (g *Game) ApplyEquipmentBonuses() {
	baseAttack := 10 + g.Player.Level*5
	baseDefense := 2 + g.Player.Level*2



	for _, equip := range g.Equipment {
		if equip.Equipped {
			baseAttack += equip.Attack
			baseDefense += equip.Defense
		}
	}

	g.Player.Attack = baseAttack
	g.Player.Defense = baseDefense
}

// Damage 计算玩家受到的伤害
func (g *Game) Damage(attack int, defense int) int {
	damage := attack - defense
	if damage < 1 {
		damage = 1
	}
	return damage
}

// Fight 战斗一次
func (g *Game) Fight() *CombatResult {
	result := &CombatResult{
		PlayerStartHP: g.Player.Health,
	}

	// 生成怪物
	monster := models.GenerateMonster(g.Player.Level)
	result.Monster = monster

	// 应用装备加成
	g.ApplyEquipmentBonuses()

	// 战斗循环
	round := 0
	for monster.IsAlive() && g.Player.Health > 0 {
		round++

		// 玩家攻击怪物
		playerDamage := g.Damage(g.Player.Attack, monster.Defense)
		monster.Health -= playerDamage
		result.PlayerDamage += playerDamage
		result.Rounds++

		if !monster.IsAlive() {
			break
		}

		// 怪物攻击玩家
		monsterDamage := g.Damage(monster.Attack, g.Player.Defense)
		g.Player.Health -= monsterDamage
		result.MonsterDamage += monsterDamage
	}

	result.PlayerEndHP = g.Player.Health
	result.Victory = !monster.IsAlive()

	if result.Victory {
		result.ExpGained = monster.ExpReward
		result.GoldGained = monster.GoldReward
		g.Player.AddExp(result.ExpGained)
		g.Player.Gold += result.GoldGained

		// 随机掉落装备
		rand.Seed(time.Now().UnixNano())
		if rand.Float64() < monster.DropRate {
			equip := models.GenerateEquipment(monster.Level)
			err := db.SaveEquipment(equip, g.Player.ID)
			if err == nil {
				equip.PlayerID = g.Player.ID
				g.Equipment = append(g.Equipment, equip)
				result.DroppedEquipment = equip
			}
		}
	} else {
		// 战败，恢复部分生命
		g.Player.Health = g.Player.MaxHealth / 2
	}

	// 更新玩家数据数据
	g.Player.LastFight = sql.NullTime{Time: time.Now(), Valid: true}
	err := db.UpdatePlayer(g.Player)
	if err != nil {
		log.Printf("Warning: failed to update player: %v", err)
	}

	return result
}

// CombatResult 战斗结果
type CombatResult struct {
	Monster          *models.Monster
	PlayerStartHP    int
	PlayerEndHP      int
	PlayerDamage     int
	MonsterDamage    int
	Rounds           int
	Victory          bool
	ExpGained        int64
	GoldGained       int64
	DroppedEquipment *models.Equipment
}

// Display 显示战斗结果
func (cr *CombatResult) Display() string {
	status := "胜利!"
	if !cr.Victory {
		status = "战败..."
	}

	result := fmt.Sprintf(
		"\n=== 战斗结果 ===\n"+
		"怪物: %s (Lv.%d)\n"+
		"回合数: %d\n"+
		"玩家: HP %d → %d\n"+
		"玩家造成伤害: %d\n"+
		"怪物造成伤害: %d\n"+
		"结果: %s",
		cr.Monster.Name, cr.Monster.Level, cr.Rounds,
		cr.PlayerStartHP, cr.PlayerEndHP, cr.PlayerDamage, cr.MonsterDamage, status,
	)

	if cr.Victory {
		result += fmt.Sprintf("\n获得: %d 经验, %d 金币", cr.ExpGained, cr.GoldGained)
		if cr.DroppedEquipment != nil {
			result += fmt.Sprintf("\n掉落: %s", cr.DroppedEquipment.String())
		}
	}

	return result
}

// StartAutoFight 开启自动战斗
func (g *Game) StartAutoFight() {
	if g.AutoTicker != nil {
		return // 已经在自动战斗
	}

	g.Player.AutoFight = true
	db.UpdatePlayer(g.Player)

	g.AutoTicker = time.NewTicker(2 * time.Second)

	go func() {
		for {
			select {
			case <-g.AutoTicker.C:
				if g.Player.Health < g.Player.MaxHealth/2 {
					// 生命值低，休息恢复
					g.Player.Health += 20
					if g.Player.Health > g.Player.MaxHealth {
						g.Player.Health = g.Player.MaxHealth
					}
					db.UpdatePlayer(g.Player)
					fmt.Println("[自动战斗] 休息中... 恢复生命")
					continue
				}

				result := g.Fight()
				fmt.Println(result.Display())

			case <-g.AutoStopped:
				g.AutoTicker.Stop()
				g.AutoTicker = nil
				g.Player.AutoFight = false
				db.UpdatePlayer(g.Player)
				return
			}
		}
	}()
}

// StopAutoFight 停止自动战斗
func (g *Game) StopAutoFight() {
	if g.AutoTicker != nil {
		g.AutoStopped <- true
	}
}

// ShowStatus 显示玩家状态
func (g *Game) ShowStatus() string {
	g.ApplyEquipmentBonuses()

	return fmt.Sprintf(
		"\n=== 玩家状态 ===\n"+
		"名称: %s\n"+
		"等级: %d\n"+
		"经验: %d/%d\n"+
		"生命: %d/%d\n"+
		"攻击: %d\n"+
		"防御: %d\n"+
		"金币: %d\n"+
		"装备数: %d\n"+
		"自动战斗: %v",
		g.Player.Name, g.Player.Level, g.Player.Exp, models.ExpToLevel(g.Player.Level+1),
		g.Player.Health, g.Player.MaxHealth, g.Player.Attack, g.Player.Defense,
		g.Player.Gold, len(g.Equipment), g.Player.AutoFight,
	)
}

// ShowInventory 显示装备栏
func (g *Game) ShowInventory() string {
	if len(g.Equipment) == 0 {
		return "\n=== 装备栏 ===\n(空)"
	}

	result := "\n=== 装备栏 ===\n"
	for i, equip := range g.Equipment {
		status := ""
		if equip.Equipped {
			status = " [已装备]"
		}
		result += fmt.Sprintf("%d. %s%s\n", i+1, equip.String(), status)
	}
	return result
}

// EquipItem 装备物品
func (g *Game) EquipItem(index int) error {
	if index < 1 || index > len(g.Equipment) {
		return fmt.Errorf("无效的装备编号")
	}

	equip := g.Equipment[index-1]

	// 查找同类装备并卸下
	for _, e := range g.Equipment {
		if e.Equipped && e.Type == equip.Type {
			e.Equipped = false
			_, err := db.DB.Exec("UPDATE equipment SET equipped = 0 WHERE id = ?", e.ID)
			if err != nil {
				log.Printf("Warning: failed to unequip: %v", err)
			}
		}
	}

	// 装备新的
	equip.Equipped = true
	_, err := db.DB.Exec("UPDATE equipment SET equipped = 1 WHERE id = ?", equip.ID)
	if err != nil {
		return fmt.Errorf("failed to equip: %w", err)
	}

	// 重新加载装备
	g.Equipment, _ = db.GetPlayerEquipment(g.Player.ID)
	return nil
}

// Rest 休息恢复生命
func (g *Game) Rest() {
	heal := g.Player.MaxHealth - g.Player.Health
	g.Player.Health = g.Player.MaxHealth
	db.UpdatePlayer(g.Player)
	fmt.Printf("休息恢复 %d 点生命！\n", heal)
}

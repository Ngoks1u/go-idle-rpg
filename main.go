package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"go-idle-rpg/cache"
	"go-idle-rpg/config"
	"go-idle-rpg/db"
	"go-idle-rpg/game"
)

const (
	VERSION = "0.1.0"
)

func main() {
	// 初始化 MySQL
	mysqlCfg := config.GetMySQLConfig()
	dbCfg := db.Config{
		Host:     mysqlCfg.Host,
		Port:     mysqlCfg.Port,
		User:     mysqlCfg.User,
		Password: mysqlCfg.Password,
		Database: mysqlCfg.Database,
	}
	fmt.Printf("连接 MySQL: %s@%s:%d/%sver\n", mysqlCfg.User, mysqlCfg.Host, mysqlCfg.Port, mysqlCfg.Database)
	err := db.Init(dbCfg)
	if err != nil {
		fmt.Printf("MySQL 初始化失败: %v\n", err)
		fmt.Println("请确保 MySQL 已启动且数据库已创建: CREATE DATABASE idle_rpg;")
		os.Exit(1)
	}
	defer db.Close()

	// 初始化 Redis
	redisCfg := config.GetRedisConfig()
	cacheCfg := cache.Config{
		Addr:     redisCfg.Addr,
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	}
	fmt.Printf("连接 Redis: %s (DB %d)\n", redisCfg.Addr, redisCfg.DB)
	err = cache.Init(cacheCfg)
	if err != nil {
		fmt.Printf("Redis 初始化失败: %v\n", err)
		fmt.Println("警告: 缓存功能将不可用，游戏仍可正常运行")
	}
	defer cache.Close()

	fmt.Println("═══════════════════════════════════════")
	fmt.Println("       挂机打宝 RPG v" + VERSION)
	fmt.Println("═══════════════════════════════════════")

	// 获取玩家名称
	playerName := getPlayerName()

	// 创建游戏实例
	g, err := game.NewGame(playerName)
	if err != nil {
		fmt.Printf("创建游戏失败: %v\n", err)
		os.Exit(1)
	}

	// 显示欢迎信息
	fmt.Printf("\n欢迎回来，%s！\n", g.Player.Name)
	fmt.Println(g.ShowStatus())

	// 主循环
	runGameLoop(g)
}

func getPlayerName() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\n请输入角色名称: ")
	name, _ := reader.ReadString('\n')
	return strings.TrimSpace(name)
}

func runGameLoop(g *game.Game) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n───────────────────────────────────────")
		fmt.Println("指令列表:")
		fmt.Println("  1. 战斗 (f)")
		fmt.Println("  2. 自动战斗 (a) - 开启/关闭")
		fmt.Println("  3. 休息 (r) - 恢复生命")
		fmt.Println("  4. 状态 (s)")
		fmt.Println("  5. 装备栏 (i)")
		fmt.Println("  6. 装备 (e <编号>)")
		fmt.Println("  7. 帮助 (h)")
		fmt.Println("  8. 退出 (q)")
		fmt.Println("───────────────────────────────────────")
		fmt.Print("> ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		switch input {
		case "1", "f", "fight":
			if g.Player.Health < 10 {
				fmt.Println("生命值过低，请先休息！")
				continue
			}
			result := g.Fight()
			fmt.Println(result.Display())

		case "2", "a", "auto":
			if g.Player.AutoFight {
				fmt.Println("自动战斗已停止")
				g.StopAutoFight()
			} else {
				fmt.Println("自动战斗已开启！每2秒自动战斗一次")
				g.StartAutoFight()
			}

		case "3", "r", "rest":
			g.Rest()
			fmt.Println(g.ShowStatus())

		case "4", "s", "status":
			fmt.Println(g.ShowStatus())

		case "5", "i", "inventory":
			fmt.Println(g.ShowInventory())

		case "6", "e", "equip":
			fmt.Print("装备编号: ")
			numInput, _ := reader.ReadString('\n')
			numInput = strings.TrimSpace(numInput)
			var num int
			fmt.Sscanf(numInput, "%d", &num)
			if err := g.EquipItem(num); err != nil {
				fmt.Printf("装备失败: %v\n", err)
			} else {
				fmt.Println("装备成功！")
				fmt.Println(g.ShowStatus())
			}

		case "7", "h", "help":
			showHelp()

		case "8", "q", "quit", "exit":
			fmt.Println("再见！")
			g.StopAutoFight()
			return

		default:
			// 检查是否是 "e 数字" 格式
			if strings.HasPrefix(input, "e ") {
				numStr := strings.TrimPrefix(input, "e ")
				var num int
				fmt.Sscanf(" "+numStr, "%d", &num)
				if err := g.EquipItem(num); err != nil {
					fmt.Printf("装备失败: %v\n", err)
				} else {
					fmt.Println("装备成功！")
					fmt.Println(g.ShowStatus())
				}
			} else {
				fmt.Println("未知指令，输入 h 查看帮助")
			}
		}
	}
}

func showHelp() {
	fmt.Println(`
═══════════════════════════════════════
              游戏帮助
═══════════════════════════════════════

挂机打宝 RPG 是一个自动化战斗游戏。

基本玩法：
1. 战斗 - 与怪物战斗，获得经验和装备
2. 自动战斗 - 开启后自动持续战斗
3. 装备系统 - 怪物会掉落装备，穿戴可提升属性
4. 升级系统 - 积累经验可提升等级

怪物类型：
  • 史莱姆 - 低级怪物，易击杀
  • 哥布林 - 中级怪物，掉落不错
  • 兽人   - 高级怪物，战力较强

  • 龙     - 稀有怪物，极高收益

装备品质：
  • 普通 (白)
  • 优秀 (绿)
  • 稀有 (蓝)
  • 史诗 (紫)
  • 传说 (橙)

═══════════════════════════════════════
`)
}

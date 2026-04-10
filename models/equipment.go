package models

import (
	"fmt"
	"math/rand"
	"time"
)

// EquipmentType 装备类型
type EquipmentType int

const (
	Weapon EquipmentType = iota // 武器
	Armor                       // 护甲
	Helmet                      // 头盔
	Boots                       // 靴子
)

// EquipmentQuality 装备品质
type EquipmentQuality int

const (
	QualityCommon EquipmentQuality = iota // 普通
	QualityUncommon                       // 优秀
	QualityRare                           // 稀有
	QualityEpic                           // 史诗
	QualityLegendary                      // 传说
)

// Equipment 装备
type Equipment struct {
	ID       int64           `db:"id"`
	Name     string          `db:"name"`
	Type     EquipmentType   `db:"type"`
	Quality  EquipmentQuality `db:"quality"`
	Level    int             `db:"level"`
	Attack   int             `db:"attack"`
	Defense  int             `db:"defense"`
	PlayerID int64           `db:"player_id"`
	Equipped bool            `db:"equipped"`
}

// QualityNames 品质名称映射
var QualityNames = map[EquipmentQuality]string{
	QualityCommon:   "普通",
	QualityUncommon: "优秀",
	QualityRare:     "稀有",
	QualityEpic:     "史诗",
	QualityLegendary: "传说",
}

// TypeNames 类型名称映射
var TypeNames = map[EquipmentType]string{
	Weapon: "武器",
	Armor:  "护甲",
	Helmet: "头盔",
	Boots:  "靴子",
}

// GenerateEquipment 生成随机装备
func GenerateEquipment(level int) *Equipment {
	rand.Seed(time.Now().UnixNano())

	equipTypes := []EquipmentType{Weapon, Armor, Helmet, Boots}
	eType := equipTypes[rand.Intn(len(equipTypes))]

	// 根据随机数决定品质
	qualityRoll := rand.Float64()
	var quality EquipmentQuality
	switch {
	case qualityRoll < 0.5:
		quality = QualityCommon
	case qualityRoll < 0.8:
		quality = QualityUncommon
	case qualityRoll < 0.95:
		quality = QualityRare
	case qualityRoll < 0.99:
		quality = QualityEpic
	default:
		quality = QualityLegendary
	}

	// 品质加成
	qualityMultiplier := 1.0 + float64(quality)*0.3

	var name string
	var baseAttack, baseDefense int

	switch eType {
	case Weapon:
		weaponNames := []string{"铁剑", "钢刀", "战斧", "长枪", "魔剑"}
		name = weaponNames[rand.Intn(len(weaponNames))]
		baseAttack = 10 + level*3
	case Armor:
		armorNames := []string{"皮甲", "锁甲", "板甲", "龙鳞甲", "秘银甲"}
		name = armorNames[rand.Intn(len(armorNames))]
		baseDefense = 5 + level*2
	case Helmet:
		helmetNames := []string{"皮帽", "铁盔", "战盔", "龙骨盔", "王冠"}
		name = helmetNames[rand.Intn(len(helmetNames))]
		baseDefense = 3 + level
	case Boots:
		bootsNames := []string{"草鞋", "皮靴", "铁靴", "战靴", "神行靴"}
		name = bootsNames[rand.Intn(len(bootsNames))]
		baseDefense = 2 + level
	}

	attack := int(float64(baseAttack) * qualityMultiplier)
	defense := int(float64(baseDefense) * qualityMultiplier)

	fullName := fmt.Sprintf("%s(%s)", name, QualityNames[quality])

	return &Equipment{
		Name:     fullName,
		Type:     eType,
		Quality:  quality,
		Level:    level,
		Attack:   attack,
		Defense:  defense,
		Equipped: false,
	}
}

// String 返回装备的字符串表示
func (e *Equipment) String() string {
	stats := ""
	if e.Attack > 0 {
		stats += fmt.Sprintf(" 攻击+%d", e.Attack)
	}
	if e.Defense > 0 {
		stats += fmt.Sprintf(" 防御+%d", e.Defense)
	}
	return fmt.Sprintf("[%s] %s (Lv.%d)%s", TypeNames[e.Type], e.Name, e.Level, stats)
}

-- MySQL 初始化脚本
-- 数据库会自动创建，这里可以添加一些初始化数据

USE idle_rpg;

-- 设置字符集
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 玩家表会由应用自动创建
-- equipment 表会由应用自动创建

-- 可以在这里插入一些初始数据
-- INSERT INTO players (name, level, exp, health, max_health, attack, defense, gold)
-- VALUES ('NPC战士', 5, 500, 150, 150, 35, 12, 100);

SET FOREIGN_KEY_CHECKS = 1;

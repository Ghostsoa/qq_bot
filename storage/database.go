package storage

import (
	"fmt"
	"qq_bot/config"
	"qq_bot/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase 初始化数据库
func InitDatabase(cfg *config.DatabaseConfig) error {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 静默模式，避免SQL日志干扰
	})

	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	// 自动迁移表结构
	if err := DB.AutoMigrate(&ChatHistory{}); err != nil {
		return fmt.Errorf("数据库迁移失败: %v", err)
	}

	utils.Info("数据库连接成功")
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

package config

import (
	"encoding/json"
	"os"
)

// Config 总配置
type Config struct {
	NapCat     *NapCatConfig   `json:"napcat"`
	AI         *AIConfig       `json:"ai"`
	Database   *DatabaseConfig `json:"database"`
	AllowedQQs []int64         `json:"allowed_qqs"` // 允许使用的QQ号白名单
}

// NapCatConfig NapCat连接配置
type NapCatConfig struct {
	Host              string `json:"host"`               // 主机地址
	Port              int    `json:"port"`               // 端口
	Token             string `json:"token"`              // 访问令牌
	HeartbeatInterval int    `json:"heartbeat_interval"` // 心跳间隔(ms)
	MessageFormat     string `json:"message_format"`     // 消息格式 array/string
}

// AIConfig AI模型配置
type AIConfig struct {
	BaseURL     string  `json:"base_url"`    // API基础地址
	APIKey      string  `json:"api_key"`     // API密钥
	Model       string  `json:"model"`       // 模型名称
	MaxTokens   int     `json:"max_tokens"`  // 最大token数
	Temperature float64 `json:"temperature"` // 温度参数
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `json:"host"`     // 数据库主机
	Port     int    `json:"port"`     // 数据库端口
	User     string `json:"user"`     // 数据库用户名
	Password string `json:"password"` // 数据库密码
	DBName   string `json:"dbname"`   // 数据库名
	SSLMode  string `json:"sslmode"`  // SSL模式
}

var globalConfig *Config

// Load 加载配置文件
func Load(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	globalConfig = &cfg
	return nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

// GetDefault 获取默认配置
func GetDefault() *Config {
	return &Config{
		NapCat: &NapCatConfig{
			Host:              "127.0.0.1",
			Port:              3001,
			Token:             "X9?Hl=AJnpjWM(bw",
			HeartbeatInterval: 30000,
			MessageFormat:     "array",
		},
		AI: &AIConfig{
			BaseURL:     "https://api.deepseek.com",
			APIKey:      "sk-593692de98614e81baf15878043c30c9",
			Model:       "deepseek-chat",
			MaxTokens:   500,
			Temperature: 0.95,
		},
		Database: &DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "qq_bot",
			Password: "your_password",
			DBName:   "qq_bot_db",
			SSLMode:  "disable",
		},
		AllowedQQs: []int64{}, // 默认空，需要手动添加QQ号
	}
}

// Save 保存配置到文件
func Save(filepath string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}

// LoadSystemPrompt 从文件加载系统提示词
func LoadSystemPrompt(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

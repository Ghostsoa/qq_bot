package config

import (
	"encoding/json"
	"os"
)

// Config 总配置
type Config struct {
	NapCat *NapCatConfig `json:"napcat"`
	AI     *AIConfig     `json:"ai"`
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
	BaseURL      string  `json:"base_url"`      // API基础地址
	APIKey       string  `json:"api_key"`       // API密钥
	Model        string  `json:"model"`         // 模型名称
	MaxTokens    int     `json:"max_tokens"`    // 最大token数
	Temperature  float64 `json:"temperature"`   // 温度参数
	SystemPrompt string  `json:"system_prompt"` // 系统提示词
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
			BaseURL:      "https://api.openai.com/v1",
			APIKey:       "",
			Model:        "gpt-3.5-turbo",
			MaxTokens:    2000,
			Temperature:  0.7,
			SystemPrompt: "你是一个友好的AI助手。",
		},
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

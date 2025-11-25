package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"qq_bot/config"
	"qq_bot/utils"
)

// AIService AI服务接口
type AIService interface {
	Chat(userMessage string) (string, error)
	ChatWithHistory(messages []Message) (string, error)
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// OpenAIService OpenAI兼容服务
type OpenAIService struct {
	config *config.AIConfig
	client *http.Client
}

// ChatRequest OpenAI聊天请求
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// ChatResponse OpenAI聊天响应
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewOpenAIService 创建OpenAI服务
func NewOpenAIService(cfg *config.AIConfig) *OpenAIService {
	return &OpenAIService{
		config: cfg,
		client: &http.Client{},
	}
}

// Chat 单轮对话
func (s *OpenAIService) Chat(userMessage string) (string, error) {
	messages := []Message{
		{
			Role:    "system",
			Content: s.config.SystemPrompt,
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}
	return s.ChatWithHistory(messages)
}

// ChatWithHistory 带历史记录的对话
func (s *OpenAIService) ChatWithHistory(messages []Message) (string, error) {
	req := ChatRequest{
		Model:       s.config.Model,
		Messages:    messages,
		MaxTokens:   s.config.MaxTokens,
		Temperature: s.config.Temperature,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", s.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	utils.Debug("发送AI请求: %s", string(jsonData))

	// 发送请求
	resp, err := s.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API返回错误: %d, %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v, 原始数据: %s", err, string(body))
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("AI返回空响应")
	}

	utils.Debug("AI回复: %s", chatResp.Choices[0].Message.Content)

	return chatResp.Choices[0].Message.Content, nil
}

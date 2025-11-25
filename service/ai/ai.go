package ai

import (
	"context"
	"fmt"
	"qq_bot/config"
	"qq_bot/utils"

	"github.com/sashabaranov/go-openai"
)

// AIService AI服务接口
type AIService interface {
	ChatWithHistory(messages []openai.ChatCompletionMessage) (string, error)
}

// OpenAIService OpenAI兼容服务
type OpenAIService struct {
	config *config.AIConfig
	client *openai.Client
}

// NewOpenAIService 创建OpenAI服务
func NewOpenAIService(cfg *config.AIConfig) *OpenAIService {
	// 配置客户端
	clientConfig := openai.DefaultConfig(cfg.APIKey)
	clientConfig.BaseURL = cfg.BaseURL + "/v1" // DeepSeek需要 /v1 后缀

	client := openai.NewClientWithConfig(clientConfig)

	return &OpenAIService{
		config: cfg,
		client: client,
	}
}

// ChatWithHistory 带历史记录的对话
func (s *OpenAIService) ChatWithHistory(messages []openai.ChatCompletionMessage) (string, error) {
	req := openai.ChatCompletionRequest{
		Model:       s.config.Model,
		Messages:    messages,
		MaxTokens:   s.config.MaxTokens,
		Temperature: float32(s.config.Temperature),
	}

	utils.Debug("发送AI请求: %d条消息", len(messages))

	resp, err := s.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("AI请求失败: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("AI返回空响应")
	}

	reply := resp.Choices[0].Message.Content
	utils.Debug("AI回复: %s", reply)

	return reply, nil
}

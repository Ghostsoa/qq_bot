package message

import (
	"fmt"
	"qq_bot/protocol"
	"qq_bot/service/ai"
	"qq_bot/service/history"
	"qq_bot/service/user"
	"qq_bot/utils"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// MessageService 消息服务
type MessageService struct {
	api            *protocol.API
	aiService      ai.AIService
	userService    *user.UserService
	historyService *history.HistoryService
	systemPrompt   string
}

// NewMessageService 创建消息服务
func NewMessageService(api *protocol.API, aiService ai.AIService, systemPrompt string, allowedQQs []int64) *MessageService {
	return &MessageService{
		api:            api,
		aiService:      aiService,
		userService:    user.NewUserService(allowedQQs),
		historyService: history.NewHistoryService(),
		systemPrompt:   systemPrompt,
	}
}

// HandleMessage 处理消息事件
func (s *MessageService) HandleMessage(event *protocol.Event) {
	if event.PostType != "message" {
		return
	}

	// 获取消息文本
	msgText := event.RawMessage
	if msgText == "" {
		return
	}

	userName := getUserName(event)
	utils.Info("收到消息: [%s] %s(%d): %s", event.MessageType, userName, event.UserID, msgText)

	// 检查QQ号是否在白名单中
	if !s.userService.CheckPermission(event.UserID) {
		utils.Debug("QQ号 %d 不在白名单中，忽略消息", event.UserID)
		return // 直接忽略，不做任何回应
	}

	// 处理命令
	if strings.HasPrefix(msgText, "/") {
		s.handleCommand(event, msgText)
		return
	}

	// 权限已在上面检查过，这里不需要再检查

	// 普通消息，调用AI回复
	s.handleAIChat(event, msgText)
}

// handleCommand 处理命令
func (s *MessageService) handleCommand(event *protocol.Event, cmd string) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}

	command := parts[0]

	switch command {
	case "/help":
		s.handleHelp(event)
	case "/ping":
		s.sendReply(event, "pong!")
	case "/about":
		s.sendReply(event, "NapCat QQ机器人 v2.0\n基于Go语言开发\n支持AI对话和上下文记忆")
	case "/clear":
		s.handleClearHistory(event)
	default:
		s.sendReply(event, "未知命令: "+command+"\n输入 /help 查看可用命令")
	}
}

// handleHelp 处理帮助命令
func (s *MessageService) handleHelp(event *protocol.Event) {
	help := "可用命令:\n"
	help += "/help - 显示帮助\n"
	help += "/ping - 测试连接\n"
	help += "/about - 关于本机器人\n"
	help += "/clear - 清空对话历史\n"

	s.sendReply(event, help)
}

// handleClearHistory 清空历史
func (s *MessageService) handleClearHistory(event *protocol.Event) {
	var groupId *int64
	if event.MessageType == "group" {
		groupId = &event.GroupID
	}

	err := s.historyService.ClearUserHistory(event.UserID, groupId)
	if err != nil {
		utils.Error("清空历史失败: %v", err)
		s.sendReply(event, "清空历史失败")
		return
	}

	s.sendReply(event, "已清空您的对话历史")
}

// handleAIChat 处理AI对话
func (s *MessageService) handleAIChat(event *protocol.Event, userMessage string) {
	if s.aiService == nil {
		utils.Debug("AI服务未配置，跳过")
		return
	}

	var groupId *int64
	if event.MessageType == "group" {
		groupId = &event.GroupID
	}

	// 保存用户消息
	err := s.historyService.SaveMessage(event.UserID, groupId, "user", userMessage)
	if err != nil {
		utils.Error("保存用户消息失败: %v", err)
	}

	// 获取历史记录
	historyMessages, err := s.historyService.GetRecentHistory(event.UserID, groupId, 20) // 获取最近20条
	if err != nil {
		utils.Error("获取历史记录失败: %v", err)
	}

	// 构建完整的消息列表（系统提示 + 历史 + 当前消息）
	messages := make([]openai.ChatCompletionMessage, 0)

	// 添加系统提示词
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "system",
		Content: s.systemPrompt,
	})

	// 添加历史记录
	messages = append(messages, historyMessages...)

	// 调用AI服务
	reply, err := s.aiService.ChatWithHistory(messages)
	if err != nil {
		utils.Error("AI服务错误: %v", err)
		s.sendReply(event, "抱歉，AI服务暂时不可用")
		return
	}

	// 保存AI回复
	err = s.historyService.SaveMessage(event.UserID, groupId, "assistant", reply)
	if err != nil {
		utils.Error("保存AI回复失败: %v", err)
	}

	// 发送回复
	s.sendReply(event, reply)
}

// sendReply 发送回复
func (s *MessageService) sendReply(event *protocol.Event, text string) {
	var err error
	var message interface{}

	// 根据消息格式构建消息
	message = protocol.BuildArrayMessage(text)

	// 根据消息类型发送
	if event.MessageType == "private" {
		err = s.api.SendPrivateMessage(event.UserID, message)
	} else if event.MessageType == "group" {
		err = s.api.SendGroupMessage(event.GroupID, message)
	}

	if err != nil {
		utils.Error("发送消息失败: %v", err)
	}
}

// getUserName 获取用户名
func getUserName(event *protocol.Event) string {
	if event.Sender != nil {
		if event.Sender.Card != "" {
			return event.Sender.Card
		}
		return event.Sender.Nickname
	}
	return fmt.Sprintf("QQ%d", event.UserID)
}

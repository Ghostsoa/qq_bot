package message

import (
	"fmt"
	"qq_bot/protocol"
	"qq_bot/service/ai"
	"qq_bot/utils"
	"strings"
)

// MessageService 消息服务
type MessageService struct {
	api       *protocol.API
	aiService ai.AIService
}

// NewMessageService 创建消息服务
func NewMessageService(api *protocol.API, aiService ai.AIService) *MessageService {
	return &MessageService{
		api:       api,
		aiService: aiService,
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

	utils.Info("收到消息: [%s] %s: %s", event.MessageType, getUserName(event), msgText)

	// 处理命令
	if strings.HasPrefix(msgText, "/") {
		s.handleCommand(event, msgText)
		return
	}

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
		s.sendReply(event, "可用命令:\n/help - 显示帮助\n/ping - 测试连接\n/about - 关于本机器人")
	case "/ping":
		s.sendReply(event, "pong!")
	case "/about":
		s.sendReply(event, "NapCat QQ机器人 v1.0\n基于Go语言开发")
	default:
		s.sendReply(event, "未知命令: "+command)
	}
}

// handleAIChat 处理AI对话
func (s *MessageService) handleAIChat(event *protocol.Event, userMessage string) {
	if s.aiService == nil {
		utils.Debug("AI服务未配置，跳过")
		return
	}

	// 调用AI服务
	reply, err := s.aiService.Chat(userMessage)
	if err != nil {
		utils.Error("AI服务错误: %v", err)
		s.sendReply(event, "抱歉，AI服务暂时不可用")
		return
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
	return fmt.Sprintf("%d", event.UserID)
}

package message

import (
	"fmt"
	"qq_bot/protocol"
	"qq_bot/service/ai"
	"qq_bot/service/history"
	"qq_bot/service/relationship"
	"qq_bot/service/user"
	"qq_bot/utils"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

// MessageService 消息服务
type MessageService struct {
	api                 *protocol.API
	aiService           ai.AIService
	userService         *user.UserService
	historyService      *history.HistoryService
	relationshipService *relationship.Service
}

// NewMessageService 创建消息服务
func NewMessageService(api *protocol.API, aiService ai.AIService, relationshipService *relationship.Service, allowedQQs []int64) *MessageService {
	return &MessageService{
		api:                 api,
		aiService:           aiService,
		userService:         user.NewUserService(allowedQQs),
		historyService:      history.NewHistoryService(),
		relationshipService: relationshipService,
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
	err := s.historyService.ClearAllHistory()
	if err != nil {
		utils.Error("清空历史失败: %v", err)
		s.sendReply(event, "清空历史失败")
		return
	}

	s.sendReply(event, "已清空所有对话历史")
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

	// 获取动态系统提示词（基于关系阶段）
	systemPrompt, err := s.relationshipService.GetStagePrompt(event.UserID, groupId)
	if err != nil {
		utils.Error("获取阶段提示词失败: %v", err)
		systemPrompt = "你是一个友好的AI助手。" // 降级默认值
	}

	// 获取历史记录
	historyMessages, err := s.historyService.GetRecentHistory(event.UserID, groupId, 200) // 获取最近200条（100轮对话）
	if err != nil {
		utils.Error("获取历史记录失败: %v", err)
	}

	// 构建完整的消息列表（系统提示 + 历史 + 当前消息）
	messages := make([]openai.ChatCompletionMessage, 0)

	// 添加系统提示词
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "system",
		Content: systemPrompt,
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

	// 评估对话并更新关系
	go func() {
		evalResult, err := s.relationshipService.EvaluateAndUpdate(event.UserID, groupId, userMessage, reply)
		if err != nil {
			utils.Error("关系评估失败: %v", err)
			return
		}

		// 输出评估结果（调试用）
		if evalResult.FamiliarityChange != 0 || evalResult.TrustChange != 0 || evalResult.IntimacyChange != 0 {
			keyMark := ""
			if evalResult.IsKeyMoment {
				keyMark = " 🔥"
			}
			utils.Debug("关系评估 [QQ=%d]: 熟悉%.1f 信任%.1f 亲密%.1f%s - %s",
				event.UserID,
				evalResult.FamiliarityChange,
				evalResult.TrustChange,
				evalResult.IntimacyChange,
				keyMark,
				evalResult.Reason)
		}
	}()

	// 发送回复
	s.sendReply(event, reply)
}

// sendReply 发送回复（支持分段）
func (s *MessageService) sendReply(event *protocol.Event, text string) {
	// 按 </> 分隔消息
	parts := strings.Split(text, "</>")

	// 清理空白
	var messages []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			messages = append(messages, part)
		}
	}

	if len(messages) == 0 {
		return
	}

	// 发送第一条消息（立即发送）
	s.sendSingleMessage(event, messages[0])

	// 发送后续消息（带延迟）
	for i := 1; i < len(messages); i++ {
		delay := s.calculateDelay(messages[i])
		time.Sleep(delay)
		s.sendSingleMessage(event, messages[i])
	}
}

// sendSingleMessage 发送单条消息
func (s *MessageService) sendSingleMessage(event *protocol.Event, text string) {
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

// calculateDelay 计算发送延迟（模拟打字速度）
func (s *MessageService) calculateDelay(text string) time.Duration {
	length := len([]rune(text))

	// 基础延迟最低1秒
	if length < 10 {
		return 1 * time.Second
	} else if length < 30 {
		return 2 * time.Second
	} else {
		return 3 * time.Second
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

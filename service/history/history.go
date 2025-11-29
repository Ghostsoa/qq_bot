package history

import (
	"qq_bot/storage"
	"time"

	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

// HistoryService 对话历史服务
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService 创建历史服务
func NewHistoryService() *HistoryService {
	return &HistoryService{
		db: storage.GetDB(),
	}
}

// SaveMessage 保存消息
func (s *HistoryService) SaveMessage(qqId int64, groupId *int64, role string, content string) error {
	return s.SaveMessageWithMetadata(qqId, groupId, role, content, nil)
}

// SaveMessageWithMetadata 保存带元数据的消息
func (s *HistoryService) SaveMessageWithMetadata(qqId int64, groupId *int64, role string, content string, metadata map[string]interface{}) error {
	history := storage.ChatHistory{
		QQId:     qqId,
		GroupId:  groupId,
		Role:     role, // user/assistant/utils
		Content:  content,
		Metadata: metadata,
	}
	return s.db.Create(&history).Error
}

// GetRecentHistory 获取最近的对话历史（限制条数）
func (s *HistoryService) GetRecentHistory(qqId int64, groupId *int64, limit int) ([]openai.ChatCompletionMessage, error) {
	var histories []storage.ChatHistory

	query := s.db.Where("qq_id = ?", qqId)
	if groupId != nil {
		query = query.Where("group_id = ?", *groupId)
	} else {
		query = query.Where("group_id IS NULL")
	}

	err := query.Order("created_at DESC").Limit(limit).Find(&histories).Error
	if err != nil {
		return nil, err
	}

	// 转换为OpenAI消息格式，并反转顺序（最老的在前）
	messages := make([]openai.ChatCompletionMessage, 0, len(histories))
	for i := len(histories) - 1; i >= 0; i-- {
		history := histories[i]
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    history.Role,
			Content: history.Content,
		})
	}

	return messages, nil
}

// CleanOldHistory 清理旧的历史记录（防止数据库膨胀）
func (s *HistoryService) CleanOldHistory(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	return s.db.Where("created_at < ?", cutoff).Delete(&storage.ChatHistory{}).Error
}

// ClearUserHistory 清空用户历史（用户主动清空）
func (s *HistoryService) ClearUserHistory(qqId int64, groupId *int64) error {
	query := s.db.Where("qq_id = ?", qqId)
	if groupId != nil {
		query = query.Where("group_id = ?", *groupId)
	} else {
		query = query.Where("group_id IS NULL")
	}
	return query.Delete(&storage.ChatHistory{}).Error
}

// ClearAllHistory 清空所有聊天记录
func (s *HistoryService) ClearAllHistory() error {
	return s.db.Exec("TRUNCATE TABLE chat_histories").Error
}

package relationship

import (
	"fmt"
	"os"
	"qq_bot/storage"
	"qq_bot/utils"

	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

// Service 关系服务
type Service struct {
	evaluator *Evaluator
	db        *gorm.DB
}

// NewService 创建关系服务
func NewService(client *openai.Client, db *gorm.DB) *Service {
	return &Service{
		evaluator: NewEvaluator(client, db),
		db:        db,
	}
}

// GetStagePrompt 获取当前阶段的系统提示词
func (s *Service) GetStagePrompt(qqId int64, groupId *int64) (string, error) {
	// 🔥 关键优化：等待该用户的评估完成（如果有正在进行的）
	lock := s.evaluator.GetUserLock(qqId)
	lock.Lock()
	lock.Unlock() // 立即释放，只是为了等待

	utils.Debug("[等待机制] QQ=%d 等待评估完成，获取最新关系状态", qqId)

	// 重新获取关系状态（确保是最新的）
	rel, err := s.evaluator.GetOrCreateRelationship(qqId, groupId)
	if err != nil {
		return "", err
	}

	// 加载基础提示词
	basePrompt, err := loadBasePrompt()
	if err != nil {
		return "", err
	}

	// 加载阶段提示词
	stagePrompt, err := s.loadStagePromptFile(rel.Stage)
	if err != nil {
		return "", err
	}

	// 注入当前分数到阶段提示词
	stagePrompt = s.injectScores(stagePrompt, rel)

	// 组合完整提示词
	fullPrompt := fmt.Sprintf("%s\n\n%s", basePrompt, stagePrompt)

	return fullPrompt, nil
}

// loadStagePromptFile 加载阶段提示词文件
func (s *Service) loadStagePromptFile(stage int) (string, error) {
	stageMap := map[int]string{
		1: "stranger",
		2: "familiar",
		3: "close",
		4: "intimate",
	}

	stageName, ok := stageMap[stage]
	if !ok {
		return "", fmt.Errorf("无效的阶段: %d", stage)
	}

	filename := fmt.Sprintf("system_prompts/stage_%d_%s.txt", stage, stageName)
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("读取阶段提示词失败: %v", err)
	}

	return string(data), nil
}

// injectScores 在提示词中注入当前分数
func (s *Service) injectScores(prompt string, rel *storage.UserRelationship) string {
	scoreInfo := fmt.Sprintf("当前分数 [熟悉%.1f 信任%.1f 亲密%.1f] - ",
		rel.Familiarity, rel.Trust, rel.Intimacy)

	// 在"系统分析："后面插入分数
	return replaceFirst(prompt, "系统分析：", "系统分析："+scoreInfo)
}

// EvaluateAndUpdate 评估对话并更新关系
func (s *Service) EvaluateAndUpdate(qqId int64, groupId *int64, userMsg, aiMsg string) (*EvaluationResult, error) {
	// 获取最近5轮历史
	history, err := s.getRecentHistory(qqId, groupId, 10) // 10条记录=5轮对话
	if err != nil {
		utils.Error("获取历史记录失败: %v", err)
		history = []storage.ChatHistory{} // 继续执行，使用空历史
	}

	// 调用评估器
	result, err := s.evaluator.Evaluate(qqId, groupId, userMsg, aiMsg, history)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// getRecentHistory 获取最近N条历史记录
func (s *Service) getRecentHistory(qqId int64, groupId *int64, limit int) ([]storage.ChatHistory, error) {
	var history []storage.ChatHistory

	query := s.db.Where("qq_id = ?", qqId)
	if groupId != nil {
		query = query.Where("group_id = ?", *groupId)
	} else {
		query = query.Where("group_id IS NULL")
	}

	err := query.Order("created_at DESC").
		Limit(limit).
		Find(&history).Error

	if err != nil {
		return nil, err
	}

	// 反转顺序（从旧到新）
	reverse(history)

	return history, nil
}

// GetRelationshipStatus 获取关系状态
func (s *Service) GetRelationshipStatus(qqId int64, groupId *int64) (*storage.UserRelationship, error) {
	return s.evaluator.GetOrCreateRelationship(qqId, groupId)
}

// loadBasePrompt 加载基础提示词
func loadBasePrompt() (string, error) {
	data, err := os.ReadFile("system_prompts/base.txt")
	if err != nil {
		return "", fmt.Errorf("读取base.txt失败: %v", err)
	}
	return string(data), nil
}

// replaceFirst 替换第一个匹配的字符串
func replaceFirst(s, old, new string) string {
	if idx := stringIndex(s, old); idx >= 0 {
		return s[:idx] + new + s[idx+len(old):]
	}
	return s
}

// stringIndex 查找子字符串位置
func stringIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// reverse 反转切片
func reverse(history []storage.ChatHistory) {
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}
}

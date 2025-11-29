package relationship

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"qq_bot/storage"
	"qq_bot/utils"
	"strings"
	"sync"

	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

// EvaluationResult 评估结果
type EvaluationResult struct {
	FamiliarityChange float64 `json:"familiarity_change"`
	TrustChange       float64 `json:"trust_change"`
	IntimacyChange    float64 `json:"intimacy_change"`
	IsKeyMoment       bool    `json:"is_key_moment"`
	Reason            string  `json:"reason"`
}

// Evaluator AI关系评估器
type Evaluator struct {
	client     *openai.Client
	db         *gorm.DB
	basePrompt string
	userLocks  sync.Map // map[int64]*sync.Mutex 每个用户的专属锁
}

// NewEvaluator 创建评估器
func NewEvaluator(client *openai.Client, db *gorm.DB) *Evaluator {
	prompt := loadEvaluatorPrompt()
	return &Evaluator{
		client:     client,
		db:         db,
		basePrompt: prompt,
	}
}

// loadEvaluatorPrompt 加载评估器提示词
func loadEvaluatorPrompt() string {
	data, err := os.ReadFile("system_prompts/evaluator.txt")
	if err != nil {
		utils.Error("加载evaluator.txt失败: %v，使用默认提示词", err)
		return "你是人际关系专家，基于生物学和心理学原理评估对话。"
	}
	return string(data)
}

// GetOrCreateRelationship 获取或创建关系记录
func (e *Evaluator) GetOrCreateRelationship(qqId int64, groupId *int64) (*storage.UserRelationship, error) {
	var rel storage.UserRelationship

	query := e.db.Where("qq_id = ?", qqId)
	if groupId != nil {
		query = query.Where("group_id = ?", *groupId)
	} else {
		query = query.Where("group_id IS NULL")
	}

	err := query.First(&rel).Error
	if err == gorm.ErrRecordNotFound {
		// 创建新记录
		rel = storage.UserRelationship{
			QQId:                qqId,
			GroupId:             groupId,
			Stage:               1,
			Familiarity:         0,
			Trust:               0,
			Intimacy:            0,
			TotalMessages:       0,
			AccumulatedCount:    0,
			EvaluationThreshold: 1, // 陌生期每次都评估
		}
		if err := e.db.Create(&rel).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &rel, nil
}

// GetUserLock 获取用户专属锁（避免同一用户并发评估）
func (e *Evaluator) GetUserLock(qqId int64) *sync.Mutex {
	lock, _ := e.userLocks.LoadOrStore(qqId, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// Evaluate 评估对话并更新关系
func (e *Evaluator) Evaluate(qqId int64, groupId *int64, userMsg, aiMsg string, recentHistory []storage.ChatHistory) (*EvaluationResult, error) {
	// 获取用户专属锁，确保同一用户的评估串行执行
	lock := e.GetUserLock(qqId)
	lock.Lock()
	defer lock.Unlock()

	utils.Debug("[评估锁] QQ=%d 获取锁成功，开始评估", qqId)

	// 获取当前关系状态
	rel, err := e.GetOrCreateRelationship(qqId, groupId)
	if err != nil {
		return nil, fmt.Errorf("获取关系状态失败: %v", err)
	}

	// 累计对话次数
	rel.AccumulatedCount++
	rel.TotalMessages++

	// 🔥 优化：判断是否达到评估阈值
	shouldEvaluate := rel.AccumulatedCount >= rel.EvaluationThreshold

	if !shouldEvaluate {
		// 未达到阈值，只更新计数，不真正评估
		utils.Debug("[评估跳过] QQ=%d 累计%d/%d次，跳过AI评估",
			qqId, rel.AccumulatedCount, rel.EvaluationThreshold)

		if err := e.db.Save(rel).Error; err != nil {
			return nil, err
		}

		// 返回空结果（表示未评估）
		return &EvaluationResult{
			FamiliarityChange: 0,
			TrustChange:       0,
			IntimacyChange:    0,
			IsKeyMoment:       false,
			Reason:            fmt.Sprintf("累计中(%d/%d)", rel.AccumulatedCount, rel.EvaluationThreshold),
		}, nil
	}

	// 达到阈值，执行真正的AI评估
	utils.Debug("[AI评估] QQ=%d 达到阈值，开始AI评估", qqId)

	// 构建评估prompt
	prompt := e.buildEvaluationPrompt(rel, recentHistory, userMsg, aiMsg)

	// 调用AI评估
	result, err := e.callAIEvaluator(prompt)
	if err != nil {
		utils.Error("AI评估失败: %v，使用默认值", err)
		// 降级到简单规则
		result = e.fallbackEvaluation(userMsg, aiMsg)
	}

	// 重置累计次数
	rel.AccumulatedCount = 0

	// 更新关系状态
	if err := e.updateRelationship(rel, result); err != nil {
		return nil, fmt.Errorf("更新关系状态失败: %v", err)
	}

	return result, nil
}

// buildEvaluationPrompt 构建评估提示词
func (e *Evaluator) buildEvaluationPrompt(rel *storage.UserRelationship, history []storage.ChatHistory, userMsg, aiMsg string) string {
	// 格式化历史对话
	historyText := formatHistory(history)

	// 阶段名称
	stageName := getStageName(rel.Stage)

	prompt := fmt.Sprintf(`%s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
【当前关系状态】
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

阶段: %s (Stage %d)
熟悉度: %.1f/100
信任度: %.1f/100
亲密度: %.1f/100
对话轮数: %d

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
【对话历史】（最近%d轮）
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

%s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
【评估任务】
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

基于以上信息和生物学原理，评估最新一轮对话对关系的影响。

输出JSON格式（仅JSON，无其他内容）:
{
  "familiarity_change": 数字（可正可负，可以是小数，根据真实影响判断），
  "trust_change": 数字（可正可负，可以是小数），
  "intimacy_change": 数字（可正可负，可以是小数），
  "is_key_moment": true/false,
  "reason": "简短分析（不超过30字）"
}

重要提示：
- 客观评估，不被当前分数锚定
- 关键时刻可以产生大幅跃升（符合多巴胺机制）
- 考虑阶段特征，但以对话质量为准
- 负面互动应给予负分`,
		e.basePrompt,
		stageName, rel.Stage,
		rel.Familiarity, rel.Trust, rel.Intimacy,
		rel.TotalMessages,
		len(history),
		historyText,
	)

	return prompt
}

// callAIEvaluator 调用AI评估器
func (e *Evaluator) callAIEvaluator(prompt string) (*EvaluationResult, error) {
	resp, err := e.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   200,
			Temperature: 0.3, // 评估用低温度
		},
	)

	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("AI返回空结果")
	}

	content := resp.Choices[0].Message.Content

	// 解析JSON
	var result EvaluationResult
	if err := parseEvaluationJSON(content, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// parseEvaluationJSON 解析评估结果JSON
func parseEvaluationJSON(content string, result *EvaluationResult) error {
	// 尝试直接解析
	if err := json.Unmarshal([]byte(content), result); err == nil {
		return nil
	}

	// 尝试提取JSON部分
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start == -1 || end == -1 || start >= end {
		return fmt.Errorf("无法从响应中提取JSON")
	}

	jsonStr := content[start : end+1]
	if err := json.Unmarshal([]byte(jsonStr), result); err != nil {
		return fmt.Errorf("解析JSON失败: %v", err)
	}

	return nil
}

// fallbackEvaluation 降级评估（简单规则）
func (e *Evaluator) fallbackEvaluation(userMsg, aiMsg string) *EvaluationResult {
	userLen := len([]rune(userMsg))

	if userLen > 20 {
		return &EvaluationResult{
			FamiliarityChange: 5,
			TrustChange:       3,
			IntimacyChange:    1,
			IsKeyMoment:       false,
			Reason:            "使用降级规则评估",
		}
	}

	return &EvaluationResult{
		FamiliarityChange: 2,
		TrustChange:       0,
		IntimacyChange:    0,
		IsKeyMoment:       false,
		Reason:            "简短对话",
	}
}

// updateRelationship 更新关系状态
func (e *Evaluator) updateRelationship(rel *storage.UserRelationship, result *EvaluationResult) error {
	// 更新分数
	rel.Familiarity += result.FamiliarityChange
	rel.Trust += result.TrustChange
	rel.Intimacy += result.IntimacyChange

	// 限制在0-100范围
	rel.Familiarity = clamp(rel.Familiarity, 0, 100)
	rel.Trust = clamp(rel.Trust, 0, 100)
	rel.Intimacy = clamp(rel.Intimacy, 0, 100)

	// 检查阶段升级
	oldStage := rel.Stage
	e.checkStageUpgrade(rel)

	// 🔥 优化：根据阶段动态调整评估阈值
	e.updateEvaluationThreshold(rel)

	// 保存到数据库
	if err := e.db.Save(rel).Error; err != nil {
		return err
	}

	// 如果升级了，输出日志
	if rel.Stage > oldStage {
		utils.Info("关系升级！QQ=%d 从阶段%d升级到阶段%d (%s)",
			rel.QQId, oldStage, rel.Stage, getStageName(rel.Stage))
	}

	return nil
}

// checkStageUpgrade 检查阶段升级
func (e *Evaluator) checkStageUpgrade(rel *storage.UserRelationship) {
	// 阶段2：熟悉期
	if rel.Stage == 1 && rel.Familiarity >= 25 && rel.Trust >= 15 {
		rel.Stage = 2
	}
	// 阶段3：亲近期
	if rel.Stage == 2 && rel.Familiarity >= 55 && rel.Trust >= 45 && rel.Intimacy >= 25 {
		rel.Stage = 3
	}
	// 阶段4：暧昧期
	if rel.Stage == 3 && rel.Familiarity >= 75 && rel.Trust >= 65 && rel.Intimacy >= 50 {
		rel.Stage = 4
	}
}

// updateEvaluationThreshold 根据阶段动态调整评估阈值
func (e *Evaluator) updateEvaluationThreshold(rel *storage.UserRelationship) {
	// 根据关系阶段设置评估频率
	// 陌生期：每次都评估（threshold=1）
	// 熟悉期：每2次评估一次（threshold=2）
	// 亲近期：每3次评估一次（threshold=3）
	// 暧昧期：每2次评估一次（threshold=2，敏感期）

	thresholds := map[int]int{
		1: 1, // 陌生期：频繁评估
		2: 2, // 熟悉期：适度评估
		3: 3, // 亲近期：放缓评估
		4: 2, // 暧昧期：敏感期，增加评估
	}

	if threshold, ok := thresholds[rel.Stage]; ok {
		if rel.EvaluationThreshold != threshold {
			utils.Debug("[阈值调整] QQ=%d Stage%d 评估阈值: %d → %d",
				rel.QQId, rel.Stage, rel.EvaluationThreshold, threshold)
			rel.EvaluationThreshold = threshold
		}
	}
}

// formatHistory 格式化历史对话
func formatHistory(history []storage.ChatHistory) string {
	if len(history) == 0 {
		return "（暂无历史对话）"
	}

	var sb strings.Builder
	round := 1
	for i := 0; i < len(history); i += 2 {
		if i+1 < len(history) {
			sb.WriteString(fmt.Sprintf("第%d轮:\n", round))
			sb.WriteString(fmt.Sprintf("  用户: %s\n", history[i].Content))
			sb.WriteString(fmt.Sprintf("  AI: %s\n", history[i+1].Content))
			round++
		}
	}

	return sb.String()
}

// getStageName 获取阶段名称
func getStageName(stage int) string {
	names := map[int]string{
		1: "陌生期",
		2: "熟悉期",
		3: "亲近期",
		4: "暧昧期",
	}
	if name, ok := names[stage]; ok {
		return name
	}
	return "未知"
}

// clamp 限制数值范围
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

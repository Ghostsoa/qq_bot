package storage

import (
	"time"
)

// 删除User表，不再需要用户注册系统

// ChatHistory 对话历史表
type ChatHistory struct {
	ID        uint                   `gorm:"primaryKey" json:"id"`
	QQId      int64                  `gorm:"index;not null" json:"qq_id"`       // QQ号
	GroupId   *int64                 `gorm:"index" json:"group_id,omitempty"`   // 群号(可选)
	Role      string                 `gorm:"size:20;not null" json:"role"`      // user/utils (支持工具调用)
	Content   string                 `gorm:"type:text;not null" json:"content"` // 消息内容
	Metadata  map[string]interface{} `gorm:"type:jsonb" json:"metadata"`        // 元数据字段
	CreatedAt time.Time              `json:"created_at"`
}

// TableName 指定表名
func (ChatHistory) TableName() string {
	return "chat_histories"
}

// UserRelationship 用户关系状态表
type UserRelationship struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	QQId                int64     `gorm:"uniqueIndex;not null" json:"qq_id"`              // QQ号（唯一）
	GroupId             *int64    `gorm:"index" json:"group_id,omitempty"`                // 群号(可选)
	Stage               int       `gorm:"default:1;not null" json:"stage"`                // 关系阶段 1-4
	Familiarity         float64   `gorm:"default:0;not null" json:"familiarity"`          // 熟悉度 0-100
	Trust               float64   `gorm:"default:0;not null" json:"trust"`                // 信任度 0-100
	Intimacy            float64   `gorm:"default:0;not null" json:"intimacy"`             // 亲密度 0-100
	TotalMessages       int       `gorm:"default:0;not null" json:"total_messages"`       // 总消息数
	AccumulatedCount    int       `gorm:"default:0;not null" json:"accumulated_count"`    // 累计对话次数（用于控制评估频率）
	EvaluationThreshold int       `gorm:"default:1;not null" json:"evaluation_threshold"` // 评估阈值（累计N次后评估）
	UpdatedAt           time.Time `json:"updated_at"`
	CreatedAt           time.Time `json:"created_at"`
}

// TableName 指定表名
func (UserRelationship) TableName() string {
	return "user_relationships"
}

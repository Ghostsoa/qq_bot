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

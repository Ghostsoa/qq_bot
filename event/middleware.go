package event

import (
	"qq_bot/protocol"
	"qq_bot/utils"
	"time"
)

// LoggerMiddleware 日志中间件
func LoggerMiddleware(event *protocol.Event, next HandlerFunc) HandlerFunc {
	return func(e *protocol.Event) {
		start := time.Now()
		utils.Info("处理事件开始: %s", e.PostType)
		next(e)
		utils.Info("处理事件结束: %s, 耗时: %v", e.PostType, time.Since(start))
	}
}

// RecoverMiddleware 恢复中间件（防止panic）
func RecoverMiddleware(event *protocol.Event, next HandlerFunc) HandlerFunc {
	return func(e *protocol.Event) {
		defer func() {
			if err := recover(); err != nil {
				utils.Error("处理事件时发生panic: %v", err)
			}
		}()
		next(e)
	}
}

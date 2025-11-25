package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger 日志记录器
type Logger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	debugLog *log.Logger
}

var defaultLogger *Logger

func init() {
	defaultLogger = NewLogger()
}

// NewLogger 创建新的日志记录器
func NewLogger() *Logger {
	return &Logger{
		infoLog:  log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLog: log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLog: log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info 信息日志
func Info(format string, v ...interface{}) {
	defaultLogger.infoLog.Output(2, fmt.Sprintf(format, v...))
}

// Error 错误日志
func Error(format string, v ...interface{}) {
	defaultLogger.errorLog.Output(2, fmt.Sprintf(format, v...))
}

// Debug 调试日志
func Debug(format string, v ...interface{}) {
	defaultLogger.debugLog.Output(2, fmt.Sprintf(format, v...))
}

// LogEvent 记录事件
func LogEvent(eventType string, content string) {
	Info("[%s] %s", eventType, content)
}

// GetTimeStamp 获取时间戳
func GetTimeStamp() int64 {
	return time.Now().Unix()
}

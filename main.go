package main

import (
	"os"
	"os/signal"
	"qq_bot/config"
	"qq_bot/connection"
	"qq_bot/event"
	"qq_bot/protocol"
	"qq_bot/service/ai"
	"qq_bot/service/message"
	"qq_bot/utils"
	"syscall"
)

func main() {
	utils.Info("=== NapCat QQ机器人启动 ===")

	// 加载配置
	cfg := loadConfig()

	// 创建AI服务
	aiService := ai.NewOpenAIService(cfg.AI)
	utils.Info("AI服务初始化完成: Model=%s", cfg.AI.Model)

	// 创建事件分发器
	dispatcher := event.NewDispatcher()

	// 注册中间件
	dispatcher.Use(event.RecoverMiddleware)
	dispatcher.Use(event.LoggerMiddleware)

	// 创建WebSocket客户端
	wsClient := connection.NewWSClient(cfg.NapCat, dispatcher.Dispatch)

	// 创建协议API
	api := protocol.NewAPI(wsClient.SendMessage)

	// 创建消息服务
	msgService := message.NewMessageService(api, aiService)

	// 注册事件处理器
	dispatcher.OnMessage(msgService.HandleMessage)
	dispatcher.OnMeta(handleMetaEvent)

	// 启动WebSocket连接
	if err := wsClient.Start(); err != nil {
		utils.Error("启动WebSocket失败: %v", err)
		os.Exit(1)
	}

	utils.Info("机器人启动成功，等待事件...")

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	utils.Info("正在关闭机器人...")
	wsClient.Stop()
	utils.Info("机器人已关闭")
}

// loadConfig 加载配置
func loadConfig() *config.Config {
	configFile := "config.json"

	// 检查配置文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		utils.Info("配置文件不存在，创建默认配置: %s", configFile)
		defaultCfg := config.GetDefault()
		if err := config.Save(configFile, defaultCfg); err != nil {
			utils.Error("保存默认配置失败: %v", err)
			os.Exit(1)
		}
		return defaultCfg
	}

	// 加载配置文件
	if err := config.Load(configFile); err != nil {
		utils.Error("加载配置文件失败: %v", err)
		os.Exit(1)
	}

	cfg := config.Get()
	utils.Info("配置加载成功")
	utils.Info("NapCat服务器: ws://%s:%d", cfg.NapCat.Host, cfg.NapCat.Port)

	return cfg
}

// handleMetaEvent 处理元事件
func handleMetaEvent(e *protocol.Event) {
	if e.MetaEventType == "heartbeat" {
		utils.Debug("收到心跳: interval=%d", e.Interval)
	} else if e.MetaEventType == "lifecycle" {
		utils.Info("生命周期事件: %s", e.SubType)
	}
}

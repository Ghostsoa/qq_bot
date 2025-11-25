package connection

import (
	"encoding/json"
	"fmt"
	"net/url"
	"qq_bot/config"
	"qq_bot/protocol"
	"qq_bot/utils"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSClient WebSocket客户端
type WSClient struct {
	conn              *websocket.Conn
	config            *config.NapCatConfig
	messageHandler    func(*protocol.Event)
	reconnectInterval time.Duration
	mu                sync.Mutex
	isRunning         bool
	stopChan          chan struct{}
}

// NewWSClient 创建WebSocket客户端
func NewWSClient(cfg *config.NapCatConfig, handler func(*protocol.Event)) *WSClient {
	return &WSClient{
		config:            cfg,
		messageHandler:    handler,
		reconnectInterval: 5 * time.Second,
		stopChan:          make(chan struct{}),
	}
}

// Connect 连接到WebSocket服务器
func (c *WSClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", c.config.Host, c.config.Port),
		Path:   "/",
	}

	// 添加token到请求头
	header := make(map[string][]string)
	if c.config.Token != "" {
		header["Authorization"] = []string{"Bearer " + c.config.Token}
	}

	utils.Info("正在连接到 %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}

	c.conn = conn
	c.isRunning = true
	utils.Info("WebSocket连接成功")

	return nil
}

// Start 启动客户端（包含自动重连）
func (c *WSClient) Start() error {
	if err := c.Connect(); err != nil {
		return err
	}

	go c.heartbeat()
	go c.readMessages()

	return nil
}

// readMessages 读取消息
func (c *WSClient) readMessages() {
	defer func() {
		c.mu.Lock()
		c.isRunning = false
		c.mu.Unlock()
	}()

	for {
		select {
		case <-c.stopChan:
			utils.Info("停止读取消息")
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				utils.Error("读取消息错误: %v", err)
				// 尝试重连
				c.reconnect()
				return
			}

			// 解析事件
			var event protocol.Event
			if err := json.Unmarshal(message, &event); err != nil {
				utils.Error("解析事件错误: %v, 原始数据: %s", err, string(message))
				continue
			}

			// 处理事件
			if c.messageHandler != nil {
				go c.messageHandler(&event)
			}
		}
	}
}

// SendMessage 发送消息
func (c *WSClient) SendMessage(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("websocket未连接")
	}

	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// heartbeat 心跳保持
func (c *WSClient) heartbeat() {
	ticker := time.NewTicker(time.Duration(c.config.HeartbeatInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.mu.Lock()
			if c.conn != nil {
				// 发送心跳包
				err := c.conn.WriteMessage(websocket.PingMessage, []byte{})
				if err != nil {
					utils.Error("心跳发送失败: %v", err)
				}
			}
			c.mu.Unlock()
		}
	}
}

// reconnect 重连
func (c *WSClient) reconnect() {
	utils.Info("尝试重新连接...")

	for {
		select {
		case <-c.stopChan:
			return
		default:
			time.Sleep(c.reconnectInterval)

			if err := c.Connect(); err != nil {
				utils.Error("重连失败: %v", err)
				continue
			}

			utils.Info("重连成功")
			go c.readMessages()
			return
		}
	}
}

// Stop 停止客户端
func (c *WSClient) Stop() error {
	close(c.stopChan)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsRunning 检查是否运行中
func (c *WSClient) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isRunning
}

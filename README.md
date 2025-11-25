# NapCat QQ 机器人

基于 Go 语言开发的 QQ 机器人，使用 NapCat 连接 QQ，支持 AI 对话功能。

## 项目架构

采用分层架构设计：

```
qq_bot/
├── config/           # 配置层 - 管理机器人和AI配置
├── connection/       # 连接层 - WebSocket客户端封装
├── protocol/         # 协议层 - OneBot协议消息结构
├── event/            # 事件层 - 事件路由和分发
├── service/          # 服务层 - 业务逻辑
│   ├── ai/          # AI服务（支持OpenAI格式）
│   └── message/     # 消息处理服务
└── utils/           # 工具层 - 日志等工具
```

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置

首次运行会自动生成 `config.json`，需要修改以下配置：

- **NapCat 配置**：已预设为 `127.0.0.1:3001`
- **AI 配置**：修改 `api_key` 和 `base_url`

配置示例：
```json
{
  "napcat": {
    "host": "127.0.0.1",
    "port": 3001,
    "token": "X9?Hl=AJnpjWM(bw",
    "heartbeat_interval": 30000,
    "message_format": "array"
  },
  "ai": {
    "base_url": "https://api.openai.com/v1",
    "api_key": "your_api_key_here",
    "model": "gpt-3.5-turbo",
    "max_tokens": 2000,
    "temperature": 0.7,
    "system_prompt": "你是一个友好的AI助手。"
  }
}
```

### 3. 运行

```bash
go run main.go
```

## 功能特性

### 基础功能
- ✅ WebSocket 连接到 NapCat
- ✅ 自动重连和心跳保持
- ✅ 事件分发和中间件支持
- ✅ 消息接收和发送

### AI 对话
- ✅ 支持 OpenAI 格式的大模型
- ✅ 可自定义系统提示词
- ✅ 支持温度、最大 Token 等参数

### 命令系统
- `/help` - 显示帮助信息
- `/ping` - 测试连接
- `/about` - 关于本机器人

## 扩展开发

### 添加新的服务

在 `service/` 目录下创建新的服务模块，例如：

```go
// service/weather/weather.go
package weather

type WeatherService struct {}

func NewWeatherService() *WeatherService {
    return &WeatherService{}
}

func (s *WeatherService) GetWeather(city string) (string, error) {
    // 实现天气查询逻辑
}
```

### 添加新的事件处理器

在 `main.go` 中注册：

```go
dispatcher.OnNotice(handleNoticeEvent)
dispatcher.OnRequest(handleRequestEvent)
```

## 技术栈

- **语言**: Go 1.23+
- **WebSocket**: gorilla/websocket
- **协议**: OneBot v11
- **AI**: OpenAI API 兼容

## 许可证

MIT License

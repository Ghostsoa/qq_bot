package event

import (
	"qq_bot/protocol"
	"qq_bot/utils"
)

// HandlerFunc 事件处理函数
type HandlerFunc func(*protocol.Event)

// Dispatcher 事件分发器
type Dispatcher struct {
	messageHandlers []HandlerFunc
	noticeHandlers  []HandlerFunc
	requestHandlers []HandlerFunc
	metaHandlers    []HandlerFunc
	middlewares     []MiddlewareFunc
}

// MiddlewareFunc 中间件函数
type MiddlewareFunc func(*protocol.Event, HandlerFunc) HandlerFunc

// NewDispatcher 创建事件分发器
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		messageHandlers: make([]HandlerFunc, 0),
		noticeHandlers:  make([]HandlerFunc, 0),
		requestHandlers: make([]HandlerFunc, 0),
		metaHandlers:    make([]HandlerFunc, 0),
		middlewares:     make([]MiddlewareFunc, 0),
	}
}

// OnMessage 注册消息事件处理器
func (d *Dispatcher) OnMessage(handler HandlerFunc) {
	d.messageHandlers = append(d.messageHandlers, handler)
}

// OnNotice 注册通知事件处理器
func (d *Dispatcher) OnNotice(handler HandlerFunc) {
	d.noticeHandlers = append(d.noticeHandlers, handler)
}

// OnRequest 注册请求事件处理器
func (d *Dispatcher) OnRequest(handler HandlerFunc) {
	d.requestHandlers = append(d.requestHandlers, handler)
}

// OnMeta 注册元事件处理器
func (d *Dispatcher) OnMeta(handler HandlerFunc) {
	d.metaHandlers = append(d.metaHandlers, handler)
}

// Use 使用中间件
func (d *Dispatcher) Use(middleware MiddlewareFunc) {
	d.middlewares = append(d.middlewares, middleware)
}

// Dispatch 分发事件
func (d *Dispatcher) Dispatch(event *protocol.Event) {
	if event == nil {
		return
	}

	utils.Debug("收到事件: post_type=%s, message_type=%s, notice_type=%s",
		event.PostType, event.MessageType, event.NoticeType)

	var handlers []HandlerFunc

	// 根据事件类型选择处理器
	switch event.PostType {
	case "message":
		handlers = d.messageHandlers
	case "notice":
		handlers = d.noticeHandlers
	case "request":
		handlers = d.requestHandlers
	case "meta_event":
		handlers = d.metaHandlers
	default:
		utils.Debug("未知事件类型: %s", event.PostType)
		return
	}

	// 执行处理器
	for _, handler := range handlers {
		finalHandler := handler
		// 应用中间件（逆序）
		for i := len(d.middlewares) - 1; i >= 0; i-- {
			finalHandler = d.middlewares[i](event, finalHandler)
		}
		finalHandler(event)
	}
}

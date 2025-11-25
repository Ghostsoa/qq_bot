package protocol

import (
	"encoding/json"
	"fmt"
)

// API OneBot API封装
type API struct {
	sender func([]byte) error
}

// NewAPI 创建API实例
func NewAPI(sender func([]byte) error) *API {
	return &API{sender: sender}
}

// SendPrivateMessage 发送私聊消息
func (a *API) SendPrivateMessage(userID int64, message interface{}) error {
	req := SendMessageReq{
		Action: "send_private_msg",
		Params: map[string]interface{}{
			"user_id": userID,
			"message": message,
		},
	}
	return a.sendRequest(req)
}

// SendGroupMessage 发送群消息
func (a *API) SendGroupMessage(groupID int64, message interface{}) error {
	req := SendMessageReq{
		Action: "send_group_msg",
		Params: map[string]interface{}{
			"group_id": groupID,
			"message":  message,
		},
	}
	return a.sendRequest(req)
}

// SendMessage 发送消息（自动判断类型）
func (a *API) SendMessage(messageType string, id int64, message interface{}) error {
	if messageType == "private" {
		return a.SendPrivateMessage(id, message)
	}
	return a.SendGroupMessage(id, message)
}

// GetLoginInfo 获取登录信息
func (a *API) GetLoginInfo() error {
	req := SendMessageReq{
		Action: "get_login_info",
		Params: map[string]interface{}{},
	}
	return a.sendRequest(req)
}

// sendRequest 发送请求
func (a *API) sendRequest(req SendMessageReq) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request error: %v", err)
	}
	return a.sender(data)
}

// BuildArrayMessage 构建array格式消息
func BuildArrayMessage(text string) []Message {
	return []Message{
		{
			Type: "text",
			Data: map[string]interface{}{
				"text": text,
			},
		},
	}
}

// BuildTextMessage 构建纯文本消息
func BuildTextMessage(text string) string {
	return text
}

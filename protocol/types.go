package protocol

// Message OneBot消息结构
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Event OneBot事件结构
type Event struct {
	Time          int64       `json:"time"`
	SelfID        int64       `json:"self_id"`
	PostType      string      `json:"post_type"`
	MessageType   string      `json:"message_type,omitempty"`
	SubType       string      `json:"sub_type,omitempty"`
	MessageID     int32       `json:"message_id,omitempty"`
	UserID        int64       `json:"user_id,omitempty"`
	GroupID       int64       `json:"group_id,omitempty"`
	Message       interface{} `json:"message,omitempty"`
	RawMessage    string      `json:"raw_message,omitempty"`
	Font          int32       `json:"font,omitempty"`
	Sender        *Sender     `json:"sender,omitempty"`
	NoticeType    string      `json:"notice_type,omitempty"`
	RequestType   string      `json:"request_type,omitempty"`
	MetaEventType string      `json:"meta_event_type,omitempty"`
	Interval      int64       `json:"interval,omitempty"`
	Status        interface{} `json:"status,omitempty"`
}

// Sender 发送者信息
type Sender struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Card     string `json:"card,omitempty"`
	Role     string `json:"role,omitempty"`
}

// SendMessageReq 发送消息请求
type SendMessageReq struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
	Echo   string                 `json:"echo,omitempty"`
}

// Response API响应
type Response struct {
	Status  string      `json:"status"`
	RetCode int         `json:"retcode"`
	Data    interface{} `json:"data,omitempty"`
	Echo    string      `json:"echo,omitempty"`
}

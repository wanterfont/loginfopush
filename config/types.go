package config

// NotifierType 通知渠道类型
type NotifierType string

const (
	NotifierTypeFCM      NotifierType = "fcm"      // Firebase Cloud Messaging
	NotifierTypeTelegram NotifierType = "telegram" // Telegram Bot
	NotifierTypeBark     NotifierType = "bark"     // Bark
	NotifierTypeWeCom    NotifierType = "wecom"    // WeCom
	NotifierTypeWxPusher NotifierType = "wxpusher" // WxPusher
)

// EventType 事件类型
type EventType string

const (
	EventTypeBan     EventType = "ban"     // IP 被封禁
	EventTypeFailure EventType = "fail"    // 登录失败
	EventTypeSuccess EventType = "success" // 登录成功
)

// ServerConfig 服务器配置
type ServerConfig struct {
	Name string `json:"name"` // 服务器名称
	Tag  string `json:"tag"`  // 服务器标签
}

// NotifierConfig 通知渠道配置
type NotifierConfig struct {
	Type    NotifierType `json:"type"`    // 通知类型
	Enabled bool         `json:"enabled"` // 是否启用
	Config  interface{}  `json:"config"`  // 具体配置（不同渠道的配置不同）
}

// FCMConfig FCM 配置
type FCMConfig struct {
	WebhookURL  string `json:"webhook_url"`  // FCM Webhook URL
	DeviceToken string `json:"device_token"` // FCM 设备 Token
}

// TelegramConfig Telegram 配置
type TelegramConfig struct {
	WebhookURL string `json:"webhook_url"` // Telegram Bot API URL
	ChatID     string `json:"chat_id"`     // 聊天 ID
}

// BarkConfig Bark 配置
type BarkConfig struct {
	WebhookURL  string `json:"webhook_url"`  // Telegram Bot API URL
	DeviceToken string `json:"device_token"` // 聊天 ID
}

// WeComConfig WeCom 配置
type WeComConfig struct {
	WebhookURL string `json:"webhook_url"` // Telegram Bot API URL
	SendKey    string `json:"send_key"`    // 聊天 ID
}

// WxPusherConfig WxPusher 配置
type WxPusherConfig struct {
	AppToken string   `json:"app_token"` // WxPusher 应用 Token
	UIDs     []string `json:"uids"`      // 接收消息的用户 ID 列表
}

// EventConfig 事件配置
type EventConfig struct {
	Type      EventType `json:"type"`      // 事件类型
	Enabled   bool      `json:"enabled"`   // 是否启用
	Title     string    `json:"title"`     // 通知标题
	Template  string    `json:"template"`  // 消息模板
	Icon      string    `json:"icon"`      // 显示图标
	Notifiers []string  `json:"notifiers"` // 使用的通知渠道
}

// Config 总配置结构
type Config struct {
	Server    ServerConfig              `json:"server"`    // 服务器配置
	Notifiers map[string]NotifierConfig `json:"notifiers"` // 通知渠道配置
	Events    map[string]EventConfig    `json:"events"`    // 事件配置
}

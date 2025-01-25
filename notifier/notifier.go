package notifier

import (
	"fmt"
	"loginfopush/config"
)

// Message 消息结构
type Message struct {
	Title    string                 // 标题
	Content  string                 // 内容
	Metadata map[string]interface{} // 元数据
}

// Notifier 通知器接口
type Notifier interface {
	Send(msg Message) error
}

// notifierFactory 通知器工厂函数类型
type notifierFactory func(config interface{}) (Notifier, error)

// 注册的通知器工厂
var factories = make(map[config.NotifierType]notifierFactory)

// RegisterNotifier 注册通知器工厂
func RegisterNotifier(typ config.NotifierType, factory notifierFactory) {
	factories[typ] = factory
}

// CreateNotifier 创建通知器实例
func CreateNotifier(cfg config.NotifierConfig) (Notifier, error) {
	factory, ok := factories[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("未知的通知器类型: %s", cfg.Type)
	}

	return factory(cfg.Config)
}

// NotifierManager 通知管理器
type NotifierManager struct {
	notifiers map[string]Notifier
	config    *config.Config
}

// NewNotifierManager 创建通知管理器
func NewNotifierManager(cfg *config.Config) (*NotifierManager, error) {
	manager := &NotifierManager{
		notifiers: make(map[string]Notifier),
		config:    cfg,
	}

	// 初始化所有启用的通知器
	for name, notifierCfg := range cfg.Notifiers {
		if !notifierCfg.Enabled {
			continue
		}

		notifier, err := CreateNotifier(notifierCfg)
		if err != nil {
			return nil, fmt.Errorf("创建通知器 %s 失败: %v", name, err)
		}

		manager.notifiers[name] = notifier
	}

	return manager, nil
}

// SendEvent 发送事件通知
func (m *NotifierManager) SendEvent(eventType config.EventType, data map[string]interface{}) error {
	// 查找事件配置
	var eventConfig config.EventConfig
	for _, evt := range m.config.Events {
		if evt.Type == eventType && evt.Enabled {
			eventConfig = evt
			break
		}
	}

	if eventConfig.Type == "" {
		return fmt.Errorf("未找到事件配置或事件未启用: %s", eventType)
	}

	// 准备模板数据
	templateData := TemplateData{
		Server:   m.config.Server,
		IP:       data["IP"].(string),
		Location: data["Location"].(string),
		Time:     data["Time"].(string),
		Details:  data["Details"].(string),
		Raw:      data["Raw"].(string),
		Extra:    data,
	}

	// 渲染模板
	content, err := RenderTemplate(eventConfig.Template, templateData)
	if err != nil {
		return fmt.Errorf("渲染模板失败: %v", err)
	}

	// 构建消息
	msg := Message{
		Title:    eventConfig.Title,
		Content:  content,
		Metadata: data,
	}

	// 发送到指定的通知渠道
	var lastErr error
	for _, name := range eventConfig.Notifiers {
		if notifier, ok := m.notifiers[name]; ok {
			if err := notifier.Send(msg); err != nil {
				lastErr = fmt.Errorf("通知器 %s 发送失败: %v", name, err)
				fmt.Printf("警告: %v\n", lastErr)
			}
		}
	}

	return lastErr
}

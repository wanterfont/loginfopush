package _func

import (
	"fmt"
	"loginfopush/config"
	"loginfopush/func/monitors"
	"loginfopush/notifier"
	_ "loginfopush/notifier/bark"     // 注册 Bark 通知器
	_ "loginfopush/notifier/fcm"      // 注册 FCM 通知器
	_ "loginfopush/notifier/telegram" // 注册 Telegram 通知器
	_ "loginfopush/notifier/wecom"    // 注册 WeCom 通知器
	_ "loginfopush/notifier/wxpusher" // 注册 WxPusher 通知器
	"sync"
	"time"
)

var notifierManager *notifier.NotifierManager

// InitMonitor 初始化监控系统
func InitMonitor(cfg *config.Config) error {
	var err error
	notifierManager, err = notifier.NewNotifierManager(cfg)
	if err != nil {
		return fmt.Errorf("初始化通知管理器失败: %v", err)
	}
	return nil
}

// sendNotification 发送通知
func sendNotification(event monitors.Event) error {
	// 准备事件数据
	data := map[string]interface{}{
		"IP":       event.IP,
		"Location": event.Location,
		"Details":  event.Details,
		"Time":     time.Now().Format("2006-01-02 15:04:05"),
		"Raw":      event.Raw,
	}

	// 发送事件通知
	return notifierManager.SendEvent(config.EventType(event.Type), data)
}

// StartMonitor 启动所有日志监控
func StartMonitor() error {
	// 创建事件通道
	eventChan := make(chan monitors.Event)

	// 创建等待组
	var wg sync.WaitGroup

	// 启动所有配置的日志监控
	for _, config := range monitors.LogConfigs {
		m, err := monitors.NewLogMonitor(config)
		if err != nil {
			fmt.Printf("警告: 无法启动 %s 监控: %v\n", config.Path, err)
			continue
		}
		defer m.Close()

		wg.Add(1)
		go func(m *monitors.LogMonitor) {
			defer wg.Done()
			m.Start(eventChan)
		}(m)
	}

	// 启动事件处理
	go func() {
		for event := range eventChan {
			if err := sendNotification(event); err != nil {
				fmt.Printf("发送通知失败: %v\n", err)
			} else {
				fmt.Printf("已发送通知: %s\n", event.Details)
			}
		}
	}()

	// 等待所有监控结束
	wg.Wait()
	return nil
}

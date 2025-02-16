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
	"strings"
	"sync"
	"time"
)

var notifierManager *notifier.NotifierManager
var monitorWg sync.WaitGroup
var monitorStopChan chan struct{}

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

// scheduleRestart 调度定时重启
func scheduleRestart() {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), 12, 5, 0, 0, now.Location())
		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}

		duration := next.Sub(now)
		fmt.Printf("下次重启时间: %v (%.2f 小时后)\n", next.Format("2006-01-02 15:04:05"), duration.Hours())

		select {
		case <-time.After(duration):
			fmt.Println("执行定时重启...")
			// 通知所有监控器停止
			close(monitorStopChan)
			// 等待所有监控器停止
			monitorWg.Wait()
			// 重新创建停止通道
			monitorStopChan = make(chan struct{})
			// 重新启动监控
			go startMonitors()
		}
	}
}

// startMonitors 启动所有监控器
func startMonitors() {
	// 创建事件通道
	eventChan := make(chan monitors.Event)

	// 记录是否有任何监控器成功启动
	monitorsStarted := false

	// 启动所有配置的日志监控
	for _, config := range monitors.LogConfigs {
		m, err := monitors.NewLogMonitor(config)
		if err != nil {
			if strings.Contains(err.Error(), "均未启用") {
				fmt.Printf("跳过监控 %s: %v\n", config.Path, err)
				continue
			}
			fmt.Printf("警告: 无法启动 %s 监控: %v\n", config.Path, err)
			continue
		}
		monitorsStarted = true

		monitorWg.Add(1)
		go func(m *monitors.LogMonitor) {
			defer monitorWg.Done()
			defer m.Close() // 移动到这里，确保在监控器实际停止时关闭文件
			m.Start(eventChan, monitorStopChan)
		}(m)
	}

	// 如果没有任何监控器启动，返回错误
	if !monitorsStarted {
		fmt.Printf("未能启动任何日志监控，请检查配置和事件启用状态\n")
		return
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
}

// StartMonitor 启动监控
func StartMonitor() error {
	// 初始化停止通道
	monitorStopChan = make(chan struct{})

	// 启动定时重启协程
	go scheduleRestart()

	// 启动监控
	startMonitors()

	// 保持主程序运行
	select {}
}

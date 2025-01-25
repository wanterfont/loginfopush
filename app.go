package main

import (
	"fmt"
	"loginfopush/config"
	_func "loginfopush/func"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 加载配置文件
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化监控系统
	if err := _func.InitMonitor(cfg); err != nil {
		fmt.Printf("初始化监控系统失败: %v\n", err)
		os.Exit(1)
	}

	// 创建一个通道来接收系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动错误通道
	errChan := make(chan error)

	// 在后台启动监控
	go func() {
		errChan <- _func.StartMonitor()
	}()

	fmt.Println("F2B 监控服务已启动...")
	fmt.Println("按 Ctrl+C 停止服务")

	// 等待信号或错误
	select {
	case err := <-errChan:
		if err != nil {
			fmt.Printf("监控服务发生错误: %v\n", err)
			os.Exit(1)
		}
	case sig := <-sigChan:
		fmt.Printf("\n收到信号 %v，正在停止服务...\n", sig)
	}

	fmt.Println("服务已停止")
}

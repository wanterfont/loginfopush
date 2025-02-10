package monitors

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"loginfopush/config"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// LogMonitor 日志监控器
type LogMonitor struct {
	config LogConfig
	file   *os.File
	reader *bufio.Reader
}

// NewLogMonitor 创建新的日志监控器
func NewLogMonitor(config LogConfig) (*LogMonitor, error) {
	file, err := os.Open(config.Path)
	if err != nil {
		return nil, fmt.Errorf("error opening log file %s: %v", config.Path, err)
	}

	// 移动到文件末尾
	file.Seek(0, 2)
	reader := bufio.NewReader(file)

	return &LogMonitor{
		config: config,
		file:   file,
		reader: reader,
	}, nil
}

// Close 关闭监控器
func (m *LogMonitor) Close() error {
	return m.file.Close()
}

// Start 开始监控
func (m *LogMonitor) Start(eventChan chan<- Event) {
	fmt.Printf("开始监控日志文件: %s\n", m.config.Path)

	for {
		line, err := m.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			fmt.Printf("读取日志错误: %v\n", err)
			continue
		}

		// 处理日志行
		if event := m.processLine(line); event != nil {
			eventChan <- *event
		}
	}
}

// processLine 处理单行日志
func (m *LogMonitor) processLine(line string) *Event {
	for _, pattern := range m.config.Patterns {
		if strings.Contains(line, pattern) {
			event := &Event{
				Raw: line,
				IP:  extractIP(line),
			}
			// 根据ip 地址查询归属 https://api.ip.sb/geoip/
			location, err := getIPLocation(event.IP)
			if err != nil {
				fmt.Printf("获取IP位置失败: %v\n", err)
			}

			event.Location = location

			// 根据日志类型和模式确定事件类型
			switch m.config.Type {
			case LogTypeFail2ban:
				if strings.Contains(line, "Ban") {
					// 检查 ban 事件是否启用
					if !isEventEnabled("ban") {
						return nil
					}
					event.Type = EventTypeBan
					event.Details = fmt.Sprintf("IP %s[%s] 已被 fail2ban 封禁", event.IP, location)
				} else if strings.Contains(line, "Found") {
					// 检查 fail 事件是否启用
					if !isEventEnabled("fail") {
						return nil
					}
					event.Type = EventTypeFailure
					event.Details = fmt.Sprintf("检测到来自 IP %s[%s] 的失败登录尝试", event.IP, location)
				}
			case LogTypeAuth:
				if strings.Contains(line, "Accepted") {
					// 检查 success 事件是否启用
					if !isEventEnabled("success") {
						return nil
					}
					event.Type = EventTypeSuccess
					// 根据日志内容判断是密码登录还是密钥登录
					if strings.Contains(line, "password") {
						event.Details = fmt.Sprintf("IP %s 密码登录成功", event.IP)
					} else if strings.Contains(line, "publickey") {
						event.Details = fmt.Sprintf("IP %s[%s] 密钥登录成功", event.IP, location)
					} else {
						event.Details = fmt.Sprintf("IP %s[%s] 登录成功", event.IP, location)
					}
				}
			}

			if event.Type != "" && event.IP != "" {
				return event
			}
		}
	}
	return nil
}

// isEventEnabled 检查事件是否启用
func isEventEnabled(eventType string) bool {
	if config.GlobalConfig == nil {
		return true // 如果配置未加载，默认启用所有事件
	}

	// 在事件配置中查找对应事件
	for _, evt := range config.GlobalConfig.Events {
		if string(evt.Type) == eventType {
			return evt.Enabled
		}
	}

	return false // 如果未找到事件配置，默认不启用
}

// extractIP 从日志行中提取 IP 地址（支持 IPv4 和 IPv6）
func extractIP(line string) string {
	// IPv4 正则表达式
	ipv4Pattern := `(\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b)`
	// IPv6 正则表达式
	ipv6Pattern := `(\b([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}\b|\b([0-9a-fA-F]{1,4}:){1,7}:|\b([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}\b|\b([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}\b|\b([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}\b|\b([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}\b|\b([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}\b|\b[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})\b|\b:((:[0-9a-fA-F]{1,4}){1,7}|:)\b|\bfe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}\b|\b::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]|[1-9]?)?[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]|[1-9]?)?[0-9])\b|\b([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]|[1-9]?)?[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]|[1-9]?)?[0-9])\b)`

	// 合并 IPv4 和 IPv6 正则表达式
	ipPattern := fmt.Sprintf("%s|%s", ipv4Pattern, ipv6Pattern)

	// 编译正则表达式
	re := regexp.MustCompile(ipPattern)

	// 查找匹配的 IP 地址
	match := re.FindString(line)
	return match
}

// getIPLocation 修改后的函数，添加重试机制和错误处理
func getIPLocation(ip string) (string, error) {
	// 如果 IP 为空，直接返回
	if ip == "" {
		return "未知位置", nil
	}

	// 最大重试次数
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		// 请求超时时间增加到 5s
		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		request, err := http.NewRequest("GET", fmt.Sprintf("https://api.ip.sb/geoip/%s", ip), nil)
		if err != nil {
			return "未知位置", err
		}

		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")

		response, err := client.Do(request)
		if err != nil {
			// 如果不是最后一次重试，则等待后继续
			if i < maxRetries-1 {
				time.Sleep(time.Second * time.Duration(i+1))
				continue
			}
			return "未知位置", err
		}
		defer response.Body.Close()

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return "未知位置", err
		}

		var ipInfo map[string]interface{}
		err = json.Unmarshal(body, &ipInfo)
		if err != nil {
			return "未知位置", err
		}

		// 获取国家,城市 拼接，如果没有城市，则只返回国家
		if city, ok := ipInfo["city"]; ok {
			return fmt.Sprintf("%s-%s", ipInfo["country"].(string), city.(string)), nil
		}
		return fmt.Sprintf("%s", ipInfo["country"]), nil
	}

	return "未知位置", fmt.Errorf("获取位置信息失败，已达到最大重试次数")
}

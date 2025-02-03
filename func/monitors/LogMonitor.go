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

// extractIP 从日志行中提取 IP 地址
func extractIP(line string) string {
	parts := strings.Fields(line)
	for _, part := range parts {
		if strings.Count(part, ".") == 3 {
			isIP := true
			for _, num := range strings.Split(part, ".") {
				if len(num) == 0 || len(num) > 3 {
					isIP = false
					break
				}
			}
			if isIP {
				return part
			}
		}
	}
	return ""
}

// 根据ip 地址查询归属 https://api.ip.sb/geoip/
func getIPLocation(ip string) (string, error) {
	// 请求超时时间 2s
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	// 设置请求头
	request, err := http.NewRequest("GET", fmt.Sprintf("https://api.ip.sb/geoip/%s", ip), nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	// 解析json
	var ipInfo map[string]interface{}
	err = json.Unmarshal(body, &ipInfo)
	if err != nil {
		return "", err
	}
	// 获取国家,城市 拼接，如果没有城市，则只返回国家
	if city, ok := ipInfo["city"]; ok {
		return fmt.Sprintf("%s-%s", ipInfo["country"].(string), city.(string)), nil
	}
	return fmt.Sprintf("%s", ipInfo["country"]), nil
}

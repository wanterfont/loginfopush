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
	"sync"
	"time"
)

// LogMonitor 日志监控器
type LogMonitor struct {
	config LogConfig
	file   *os.File
	reader *bufio.Reader
	path   string // 保存文件路径
	offset int64  // 保存读取位置
}

// NewLogMonitor 创建新的日志监控器
func NewLogMonitor(config LogConfig) (*LogMonitor, error) {
	// 首先检查是否需要监控这种类型的日志
	if !shouldMonitorLogType(config.Type) {
		return nil, fmt.Errorf("日志类型 %v 的所有事件均未启用，跳过监控", config.Type)
	}

	file, err := os.Open(config.Path)
	if err != nil {
		return nil, fmt.Errorf("error opening log file %s: %v", config.Path, err)
	}

	// 移动到文件末尾
	offset, err := file.Seek(0, 2)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("error seeking file %s: %v", config.Path, err)
	}

	reader := bufio.NewReader(file)

	return &LogMonitor{
		config: config,
		file:   file,
		reader: reader,
		path:   config.Path,
		offset: offset,
	}, nil
}

// Close 关闭监控器
func (m *LogMonitor) Close() error {
	return m.file.Close()
}

// reopenFile 重新打开文件
func (m *LogMonitor) reopenFile() error {
	// 关闭现有文件
	if m.file != nil {
		m.file.Close()
	}

	// 重新打开文件
	file, err := os.Open(m.path)
	if err != nil {
		return fmt.Errorf("error reopening file: %v", err)
	}

	// 获取文件大小
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("error getting file info: %v", err)
	}

	// 如果文件大小小于之前的偏移量，说明文件被轮转了
	if info.Size() < m.offset {
		m.offset = 0
	}

	// 设置偏移量
	_, err = file.Seek(m.offset, 0)
	if err != nil {
		file.Close()
		return fmt.Errorf("error seeking file: %v", err)
	}

	m.file = file
	m.reader = bufio.NewReader(file)
	return nil
}

// Start 开始监控
func (m *LogMonitor) Start(eventChan chan<- Event, stopChan <-chan struct{}) {
	fmt.Printf("开始监控日志文件: %s\n", m.config.Path)

	for {
		select {
		case <-stopChan:
			fmt.Printf("停止监控日志文件: %s\n", m.config.Path)
			return
		default:
			line, err := m.reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					// 保存当前位置
					if m.file != nil {
						m.offset, _ = m.file.Seek(0, 1)
					}

					// 检查文件是否被轮转
					if err := m.reopenFile(); err != nil {
						fmt.Printf("重新打开文件失败: %v\n", err)
						time.Sleep(5 * time.Second)
					}

					time.Sleep(100 * time.Millisecond)
					continue
				}
				fmt.Printf("读取日志错误: %v\n", err)
				time.Sleep(1 * time.Second)
				continue
			}

			// 更新偏移量
			m.offset += int64(len(line))

			// 处理日志行
			if event := m.processLine(line); event != nil {
				eventChan <- *event
			}
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

// 使用缓存的 IP 位置信息
var (
	ipLocationCache = make(map[string]ipLocationInfo)
	ipCacheMutex    sync.RWMutex
)

type ipLocationInfo struct {
	location  string
	timestamp time.Time
}

// getIPLocation 修改后的函数，添加缓存机制
func getIPLocation(ip string) (string, error) {
	if ip == "" {
		return "未知位置", nil
	}

	// 检查缓存
	ipCacheMutex.RLock()
	if info, exists := ipLocationCache[ip]; exists {
		// 如果缓存未过期（24小时内）
		if time.Since(info.timestamp) < 24*time.Hour {
			ipCacheMutex.RUnlock()
			return info.location, nil
		}
	}
	ipCacheMutex.RUnlock()

	// 最大重试次数
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
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

		var location string
		if city, ok := ipInfo["city"]; ok {
			location = fmt.Sprintf("%s-%s", ipInfo["country"].(string), city.(string))
		} else {
			location = fmt.Sprintf("%s", ipInfo["country"])
		}

		// 更新缓存
		ipCacheMutex.Lock()
		ipLocationCache[ip] = ipLocationInfo{
			location:  location,
			timestamp: time.Now(),
		}
		ipCacheMutex.Unlock()

		return location, nil
	}

	return "未知位置", fmt.Errorf("获取位置信息失败，已达到最大重试次数")
}

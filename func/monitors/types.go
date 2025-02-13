package monitors

import "loginfopush/config"

// LogType 定义日志类型
type LogType string

const (
	// 日志类型常量
	LogTypeFail2ban LogType = "fail2ban" // fail2ban 日志
	LogTypeAuth     LogType = "auth"     // 认证日志
)

// LogConfig 日志配置结构
type LogConfig struct {
	Type     LogType  // 日志类型
	Path     string   // 日志文件路径
	Patterns []string // 匹配模式
}

// EventType 事件类型
type EventType string

const (
	EventTypeBan     EventType = "ban"     // IP 被封禁
	EventTypeFailure EventType = "fail"    // 登录失败
	EventTypeSuccess EventType = "success" // 登录成功
)

// Event 事件结构
type Event struct {
	Type     EventType // 事件类型
	IP       string    // 相关 IP
	Location string    // 相关IP 位置
	Details  string    // 详细信息
	Raw      string    // 原始日志行
}

// LogConfigs 预定义的日志配置
var LogConfigs = []LogConfig{
	{
		Type: LogTypeFail2ban,
		Path: "/var/log/fail2ban.log",
		// Path: "./fail2ban.log",
		Patterns: []string{
			"Ban",   // 封禁事件
			"Found", // 发现攻击
		},
	},
	{
		Type: LogTypeAuth,
		Path: "/var/log/auth.log", // Debian/Ubuntu 系统
		Patterns: []string{
			"Accepted password for",   // 密码登录成功
			"Accepted publickey for",  // 密钥登录成功
			"session opened for user", // 会话开启
		},
	},
	{
		Type: LogTypeAuth,
		Path: "/var/log/secure", // CentOS/RHEL 系统
		Patterns: []string{
			"Accepted password for",   // 密码登录成功
			"Accepted publickey for",  // 密钥登录成功
			"session opened for user", // 会话开启
		},
	},
}

// shouldMonitorLogType 判断是否需要监控特定类型的日志
func shouldMonitorLogType(logType LogType) bool {
	if config.GlobalConfig == nil {
		return true // 如果配置未加载，默认监控所有日志
	}

	switch logType {
	case LogTypeFail2ban:
		// 检查是否启用了 ban 或 fail 事件
		for _, evt := range config.GlobalConfig.Events {
			if (evt.Type == "ban" || evt.Type == "fail") && evt.Enabled {
				return true
			}
		}
		return false
	case LogTypeAuth:
		// 检查是否启用了 success 事件
		for _, evt := range config.GlobalConfig.Events {
			if evt.Type == "success" && evt.Enabled {
				return true
			}
		}
		return false
	default:
		return false
	}
}

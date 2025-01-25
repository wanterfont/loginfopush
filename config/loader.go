package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	// DefaultConfigPath 默认配置文件路径
	DefaultConfigPath = "config/config.json"
	// GlobalConfig 全局配置实例
	GlobalConfig *Config
)

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	// 读取配置文件
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析配置
	config := &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证并转换具体的配置类型
	for name, notifier := range config.Notifiers {
		switch notifier.Type {
		case NotifierTypeFCM:
			var fcmConfig FCMConfig
			data, _ := json.Marshal(notifier.Config)
			if err := json.Unmarshal(data, &fcmConfig); err != nil {
				return nil, fmt.Errorf("解析 FCM 配置失败: %v", err)
			}
			notifier.Config = fcmConfig
			config.Notifiers[name] = notifier
		case NotifierTypeTelegram:
			var telegramConfig TelegramConfig
			data, _ := json.Marshal(notifier.Config)
			if err := json.Unmarshal(data, &telegramConfig); err != nil {
				return nil, fmt.Errorf("解析 Telegram 配置失败: %v", err)
			}
			notifier.Config = telegramConfig
			config.Notifiers[name] = notifier
		case NotifierTypeBark:
			var barkConfig BarkConfig
			data, _ := json.Marshal(notifier.Config)
			if err := json.Unmarshal(data, &barkConfig); err != nil {
				return nil, fmt.Errorf("解析 Bark 配置失败: %v", err)
			}
			notifier.Config = barkConfig
			config.Notifiers[name] = notifier
		case NotifierTypeWeCom:
			var wecomConfig WeComConfig
			data, _ := json.Marshal(notifier.Config)
			if err := json.Unmarshal(data, &wecomConfig); err != nil {
				return nil, fmt.Errorf("解析 WeCom 配置失败: %v", err)
			}
			notifier.Config = wecomConfig
			config.Notifiers[name] = notifier
		default:
			return nil, fmt.Errorf("不支持的通知类型: %s", notifier.Type)
		}
	}

	GlobalConfig = config
	return config, nil
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, configPath string) error {
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 转换为 JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入文件
	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

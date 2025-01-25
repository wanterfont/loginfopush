package telegram

import (
	"fmt"
	"io/ioutil"
	"loginfopush/config"
	"loginfopush/notifier"
	"net/http"
	"net/url"
)

func init() {
	notifier.RegisterNotifier(config.NotifierTypeTelegram, NewTelegramNotifier)
}

// TelegramNotifier Telegram 通知器
type TelegramNotifier struct {
	config config.TelegramConfig
}

// NewTelegramNotifier 创建 Telegram 通知器
func NewTelegramNotifier(cfg interface{}) (notifier.Notifier, error) {
	telegramConfig, ok := cfg.(config.TelegramConfig)
	if !ok {
		return nil, fmt.Errorf("无效的 Telegram 配置")
	}

	return &TelegramNotifier{
		config: telegramConfig,
	}, nil
}

// Send 发送通知
func (n *TelegramNotifier) Send(msg notifier.Message) error {
	// 对消息进行 URL 编码
	encodedMessage := url.QueryEscape(msg.Content)

	// 构建请求 URL
	requestURL := fmt.Sprintf("%s?chat_id=%s&text=%s",
		n.config.WebhookURL,
		n.config.ChatID,
		encodedMessage)

	// 发送请求
	resp, err := http.Get(requestURL)
	if err != nil {
		return fmt.Errorf("send webhook error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("webhook response error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}

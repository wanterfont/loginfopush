package bark

import (
	"fmt"
	"io/ioutil"
	"loginfopush/config"
	"loginfopush/notifier"
	"net/http"
	"net/url"
)

func init() {
	notifier.RegisterNotifier(config.NotifierTypeBark, NewBarkNotifer)
}

// BarkNotifer Telegram 通知器
type BarkNotifer struct {
	config config.BarkConfig
}

// NewBarkNotifer 创建 Telegram 通知器
func NewBarkNotifer(cfg interface{}) (notifier.Notifier, error) {
	barkConfig, ok := cfg.(config.BarkConfig)
	if !ok {
		return nil, fmt.Errorf("无效的 Bark 配置")
	}

	return &BarkNotifer{
		config: barkConfig,
	}, nil
}

// Send 发送通知
func (n *BarkNotifer) Send(msg notifier.Message) error {
	// 对消息进行 URL 编码
	encodedMessage := url.QueryEscape(msg.Content)

	// 构建请求 URL, 格式: https://api.day.app/your_device_key/content

	requestURL := fmt.Sprintf("%s/%s/%s",
		n.config.WebhookURL,
		n.config.DeviceToken,
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

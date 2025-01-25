package wecom

import (
	"fmt"
	"io/ioutil"
	"loginfopush/config"
	"loginfopush/notifier"
	"net/http"
	"net/url"
)

func init() {
	notifier.RegisterNotifier(config.NotifierTypeWeCom, NewWeComNotifer)
}

// WeComNotifer Telegram 通知器
type WeComNotifer struct {
	config config.WeComConfig
}

// NewWeComNotifer 创建 Telegram 通知器
func NewWeComNotifer(cfg interface{}) (notifier.Notifier, error) {
	wecomConfig, ok := cfg.(config.WeComConfig)
	if !ok {
		return nil, fmt.Errorf("无效的 Wecom 配置")
	}

	return &WeComNotifer{
		config: wecomConfig,
	}, nil
}

// Send 发送通知
func (n *WeComNotifer) Send(msg notifier.Message) error {
	// 对消息进行 URL 编码
	encodedMessage := url.QueryEscape(msg.Content)

	// 构建请求 URL, 格式: http://wcom.net/wecomchan?sendkey={}&msg_type={}&msg={}

	requestURL := fmt.Sprintf("%s?sendkey=%s&msg_type=text&msg=%s",
		n.config.WebhookURL,
		n.config.SendKey,
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

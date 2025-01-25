package fcm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"loginfopush/config"
	"loginfopush/notifier"
	"net/http"
)

func init() {
	notifier.RegisterNotifier(config.NotifierTypeFCM, NewFCMNotifier)
}

// FCMNotifier FCM 通知器
type FCMNotifier struct {
	config config.FCMConfig
}

// NewFCMNotifier 创建 FCM 通知器
func NewFCMNotifier(cfg interface{}) (notifier.Notifier, error) {
	fcmConfig, ok := cfg.(config.FCMConfig)
	if !ok {
		return nil, fmt.Errorf("无效的 FCM 配置")
	}

	return &FCMNotifier{
		config: fcmConfig,
	}, nil
}

// FCMPayload FCM 请求负载
type FCMPayload struct {
	Data struct {
		To       string `json:"to"`
		TTL      int    `json:"ttl"`
		Priority string `json:"priority"`
		Data     struct {
			Text struct {
				Title     string `json:"title"`
				Message   string `json:"message"`
				Clipboard bool   `json:"clipboard"`
			} `json:"text"`
		} `json:"data"`
	} `json:"data"`
}

// Send 发送通知
func (n *FCMNotifier) Send(msg notifier.Message) error {
	payload := FCMPayload{}
	payload.Data.To = n.config.DeviceToken
	payload.Data.TTL = 60
	payload.Data.Priority = "high"
	payload.Data.Data.Text.Title = msg.Title
	payload.Data.Data.Text.Message = msg.Content
	payload.Data.Data.Text.Clipboard = false

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal json error: %v", err)
	}

	resp, err := http.Post(n.config.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
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

package wxpusher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"loginfopush/config"
	"loginfopush/notifier"
	"net/http"
)

const (
	wxPusherAPI = "https://wxpusher.zjiecode.com/api/send/message"
)

type wxPusherNotifier struct {
	appToken string
	uids     []string
	server   config.ServerConfig
}

type wxPusherRequest struct {
	AppToken      string   `json:"appToken"`
	Content       string   `json:"content"`
	Summary       string   `json:"summary"`
	ContentType   int      `json:"contentType"`
	UIDs          []string `json:"uids"`
	URL           string   `json:"url,omitempty"`
	VerifyPay     bool     `json:"verifyPay"`
	VerifyPayType int      `json:"verifyPayType"`
}

func init() {
	notifier.RegisterNotifier(config.NotifierTypeWxPusher, NewNotifier)
}

// NewNotifier 创建 WxPusher 通知器
func NewNotifier(cfg interface{}) (notifier.Notifier, error) {
	wxConfig, ok := cfg.(config.WxPusherConfig)
	if !ok {
		return nil, fmt.Errorf("无效的 WxPusher 配置")
	}

	if wxConfig.AppToken == "" {
		return nil, fmt.Errorf("WxPusher AppToken 不能为空")
	}

	if len(wxConfig.UIDs) == 0 {
		return nil, fmt.Errorf("WxPusher UIDs 不能为空")
	}

	// 获取全局配置中的服务器信息
	serverConfig := config.GlobalConfig.Server

	return &wxPusherNotifier{
		appToken: wxConfig.AppToken,
		uids:     wxConfig.UIDs,
		server:   serverConfig,
	}, nil
}

// Send 发送消息
func (w *wxPusherNotifier) Send(msg notifier.Message) error {
	// 构建请求体
	reqBody := wxPusherRequest{
		AppToken:      w.appToken,
		Content:       msg.Content,
		Summary:       fmt.Sprintf("%s - %s", w.server.Name, "login"),
		ContentType:   1, // HTML
		UIDs:          w.uids,
		VerifyPay:     false,
		VerifyPayType: 0,
	}

	// 转换为 JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("JSON 编码失败: %v", err)
	}

	// 发送请求
	resp, err := http.Post(wxPusherAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

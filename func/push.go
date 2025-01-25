package _func

// 定义消息推送接口，后续支持多种推送方式
type PushInterface interface {
	Send(message string) error
}

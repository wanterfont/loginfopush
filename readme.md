# loginfopush
> 通过读取`/var/log/secure`或`/var/log/auth.log` 文件，监听登录事件，并推送到指定的通知渠道。
> 



## 使用
通过一键脚本安装，并配置参数：<br/>
`curl -fsSL https://wanterfont.github.io/loginfopush/install.sh -o install.sh && bash install.sh`

前置准备:
1. 准备好自己的[消息推送渠道](使用fcm%7Cbark%7Ctelegram推送消息.md)
2. 安装 `rsyslog` 并重启

消息内容参考：
>✅ 服务器: clawcloud (🇭🇰) <br/>
> IP: 47.108.002.001 登录成功 <br/>
> 时间: 2025-01-26 22:52:03 <br/>
> 位置: China-NanJing <br/>
> 详情: IP 47.108.002.001[China-NanJing] 密钥登录成功

目前推送渠道支持：
- FCM
- TelgramBot
- Bark
- WxPusher
- 企业微信

___

## 简介
这是一个服务器安全事件通知系统，可以监控并通过多种渠道推送服务器的登录尝试、封禁等安全事件。

## 功能特点
- 多渠道通知支持

- 可配置的事件类型
  - 登录成功通知
  - 登录失败警告
  - IP 封禁提醒
- 自定义消息模板
- 灵活的通知渠道配置

## 配置说明

### 服务器配置

### 支持的通知渠道

1. **FCM**
   - 需要配置 webhook_url 和 device_token
   
2. **Telegram**
   - 需要配置 webhook_url 和 chat_id

3. **Bark**
   - 需要配置 webhook_url 和 device_token

4. **企业微信**
   - 需要配置 webhook_url 和 send_key
5. **WxPuser**
   - 需要配置 app_token 和 Uid

### 事件类型

1. **封禁通知 (ban)**
   - 当 IP 被封禁时触发
   - 默认图标: 🚫

2. **登录失败通知 (login_failure)**
   - 当登录失败时触发
   - 默认图标: ⚠️

3. **登录成功通知 (login_success)**
   - 当登录成功时触发
   - 默认图标: ✅

## 配置示例
请参考 `config/config.json` 文件进行配置。

## 消息模板变量
所有事件消息模板支持以下变量：
- `{{.Server.Name}}`: 服务器名称
- `{{.Server.Tag}}`: 服务器标签
- `{{.IP}}`: 触发事件的 IP 地址
- `{{.Time}}`: 事件发生时间
- `{{.Location}}`: IP 地理位置
- `{{.Details}}`: 详细信息

## 使用说明
通过一键脚本安装，并配置参数：<br/>
`curl -fsSL https://wanterfont.github.io/loginfopush/install.sh -o install.sh && bash install.sh`

## 注意事项
- 请妥善保管各通知渠道的 token 和密钥
- 建议定期检查通知渠道的可用性
- 可以针对不同事件配置不同的通知渠道组合

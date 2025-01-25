# 服务器安全通知系统

## 简介
这是一个服务器安全事件通知系统，可以监控并通过多种渠道推送服务器的登录尝试、封禁等安全事件。

## 功能特点
- 多渠道通知支持
  - Firebase Cloud Messaging (FCM)
  - Telegram
  - Bark
  - 企业微信
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
1. 复制 `config/config.json.example` 到 `config/config.json`
2. 根据需要修改配置文件
3. 配置所需的通知渠道
4. 启用/禁用所需的事件通知

## 注意事项
- 请妥善保管各通知渠道的 token 和密钥
- 建议定期检查通知渠道的可用性
- 可以针对不同事件配置不同的通知渠道组合

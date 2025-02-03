记录下常用的几个消息推送服务，用于后续的服务登录提醒功能。

---

## FCM (Firebase Cloud Messaging)

FCM（Firebase Cloud Messaging）是 Google 提供的一种跨平台消息推送服务，主要用于 Android 应用。它通过 Google 的原生推送通道实现消息的实时传递，需要设备能够正常访问 FCM 服务。

### 使用前提

FCM 依赖于 Google 的服务，因此设备需要满足以下条件：

1. **能够稳定访问 FCM 服务**：设备需要能够正常连接 Google 的服务器。在某些地区，可能需要通过科学上网才能访问。
2. **安装 FCM Toolbox 软件**：这是一个用于测试和验证设备 FCM 推送功能的工具。

### 使用步骤

1. **安装 FCM Toolbox**：
    - 在设备上安装 [FCM Toolbox](https://play.google.com/store/apps/details?id=com.mitteloupe.fcmtoolbox) 应用。
    - 打开应用后，复制设备生成的 FCM Token。
2. **测试推送功能**：
    - 访问 [FCM Toolbox 测试平台](https://fcm-toolbox-public.firebaseapp.com/#send-text)。
    - 将复制的 FCM Token 粘贴到网页中的输入框。
    - 输入推送消息内容，点击发送。
3. **验证推送**：
    - 如果设备能够正常接收 FCM 推送，消息会显示在设备的通知栏中。
    - 如果无法接收推送，请检查设备是否能够稳定连接 FCM 服务。

### 请求示例（Text 消息）

在使用 FCM Toolbox 测试平台时，可以通过浏览器的开发者工具（F12）查看发送的请求详情。以下是一个典型的 FCM 推送请求示例：

```bash
curl -X POST "https://us-central1-fir-cloudmessaging-4e2cd.cloudfunctions.net/send>" \\
-H "Content-Type: application/json" \\
-d '{
  "data": {
    "to": "your_device_token",
    "ttl": 60,
    "priority": "high",
    "data": {
      "text": {
        "title": "测试标题",
        "message": "这是一条测试消息",
        "clipboard": false
      }
    }
  }
}'

```

### 常见问题

1. **设备无法接收推送**：
    - 检查设备是否能够稳定连接 FCM 服务。
    - 确认 FCM Token 是否正确且未过期。
    - 检查设备的通知权限是否已开启。
2. **请求地址不同**：
    - 不同地区或设备可能使用不同的请求地址，建议通过浏览器的开发者工具（F12）查看实际请求地址。
3. **消息优先级**：
    - 如果需要确保消息及时送达，请将 `priority` 设置为 `high`。

---



## Bark：iOS 平台的开源推送服务

Bark 是一款专为 iOS 设备设计的开源消息推送服务，允许用户通过简单的 HTTP 请求向自己的 iPhone 发送自定义通知。它基于苹果的 APNs（Apple Push Notification service）实现，具有高效、稳定、隐私安全的特点。Bark 支持自建服务端，适合对隐私和数据安全有较高要求的用户。

---

### 使用前提

1. **iOS 设备**：Bark 仅支持 iOS 设备，需安装 Bark 客户端。
2. **网络连接**：设备需要能够正常访问 Bark 服务端（默认使用官方服务端或自建服务端）。
3. **设备 Key**：在 Bark 客户端中注册设备后，获取唯一的设备 Key。

---

### 使用步骤

1. **安装 Bark 客户端**：
    - 在 App Store 搜索并下载 Bark 应用。
    - 打开应用后，注册设备并获取设备的唯一 Key（设备标识符）。
2. **发送推送请求**：
    - 使用 HTTP GET 或 POST 请求向 Bark 服务端发送消息。
    - 如果是自建服务端，需将服务端地址替换为你的私有服务器地址。

---

### 请求示例

以默认官方服务端通过 GET 请求发送推送消息的示例：

### 发送简单文本消息

```
https://api.day.app/your_device_key/测试标题/这是一条测试消息>

```

- **your_device_key**：替换为你的设备 Key。
- **测试标题**：推送消息的标题。
- **这是一条测试消息**：推送消息的内容。

---

### 常见问题

1. **设备无法接收推送**：
    - 检查设备是否能够正常访问 Bark 服务端。
    - 确认设备 Key 是否正确且未过期。
    - 检查设备的通知权限是否已开启。
2. **请求地址不同**：
    - 不同地区或设备可能使用不同的请求地址，建议通过浏览器的开发者工具（F12）查看实际请求地址。
3. **消息优先级**：
    - 如果需要确保消息及时送达，请将 `priority` 设置为 `high`。

---

## Telegram Bot 消息推送

---

### 使用前提

1. 能正常访问 tg。
2. **Bot Token**：通过 BotFather 创建 Bot 并获取 API Token。
3. **Chat ID**：需要获取目标用户或群组的 Chat ID，用于指定消息接收者。

---

### 使用步骤

1. **创建 Telegram Bot**：
    - 在 Telegram 中搜索并联系 `@BotFather`。
    - 使用 `/newbot` 命令创建一个新的 Bot，并按照提示设置名称和用户名。
    - 创建完成后，BotFather 会提供一个 API Token，保存好这个 Token。
2. **获取 Chat ID**：
    - 向你的 Bot 发送一条消息（例如 `/start`）。
    - 使用以下 API 请求获取 Chat ID：
        
        ```
        https://api.telegram.org/bot{你的BotToken}/getUpdates>
        
        ```
        
    - 在返回的 JSON 数据中，找到 `chat` 对象的 `id` 字段，这就是 Chat ID。
3. **发送消息**：
    - 使用 Telegram Bot API 发送消息，支持文本、图片、文件等多种格式。

---

### 请求示例

以下是一个通过 HTTP GET 请求发送文本消息的示例：

### 发送简单文本消息

```
<https://api.telegram.org/bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11/sendMessage?chat_id=123456789&text=这是一条测试消息>

```

- **123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11**：替换为你的 Bot Token。
- **123456789**：替换为目标 Chat ID。
- **这是一条测试消息**：推送消息的内容。

---

### 常见问题

1. **无法接收消息**：
    - 检查 Bot Token 和 Chat ID 是否正确。
    - 确保 Bot 没有被用户屏蔽。
    - 确认网络连接正常。
2. **消息格式问题**：
    - 如果使用 HTML 或 Markdown 格式，确保内容符合规范。
    - 特殊字符需进行 URL 编码。
3. **速率限制**：
    - Telegram 对 API 调用有速率限制，每分钟最多 30 条消息。如果需要更高频率，可以考虑使用批量发送或异步请求。

---
参考地址：<br/>
[bark](https://day.app/2021/06/barkfaq/)<br/>
[fcm-tool-box](https://fcm-toolbox-public.firebaseapp.com/#send-text)<br/>
[telegram bot](https://core.telegram.org/bots/api)<br/>

其他推送方式：<br/>
[wxpusher](https://wxpusher.zjiecode.com/docs/#/)
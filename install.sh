#!/bin/bash

# 帮助信息
show_help() {
    echo "用法: $0 [选项]"
    echo "选项:"
    echo "  -h, --help                显示帮助信息"
    echo "  --fcm-webhook URL         设置 FCM webhook URL"
    echo "  --fcm-token TOKEN         设置 FCM device token"
    echo "  --tg-webhook URL          设置 Telegram webhook URL"
    echo "  --tg-chatid ID            设置 Telegram chat ID"
    echo "  --bark-webhook URL        设置 Bark webhook URL"
    echo "  --bark-token TOKEN        设置 Bark device token"
    echo "  --wxpusher-token TOKEN    设置 WxPusher app token"
    echo "  --wxpusher-uids UID1,UID2 设置 WxPusher UIDs (用逗号分隔)"
    echo "  --server-name NAME        设置服务器名称"
    echo "  --server-tag TAG          设置服务器标签"
    echo "示例:"
    echo "$0 --fcm-webhook 'https://xxx' --fcm-token 'xxx' --server-name 'myserver'"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        --fcm-webhook)
            FCM_WEBHOOK="$2"
            shift 2
            ;;
        --fcm-token)
            FCM_TOKEN="$2"
            shift 2
            ;;
        --tg-webhook)
            TG_WEBHOOK="$2"
            shift 2
            ;;
        --tg-chatid)
            TG_CHATID="$2"
            shift 2
            ;;
        --bark-webhook)
            BARK_WEBHOOK="$2"
            shift 2
            ;;
        --bark-token)
            BARK_TOKEN="$2"
            shift 2
            ;;
        --wxpusher-token)
            WXPUSHER_TOKEN="$2"
            shift 2
            ;;
        --wxpusher-uids)
            WXPUSHER_UIDS="$2"
            shift 2
            ;;
        --server-name)
            SERVER_NAME="$2"
            shift 2
            ;;
        --server-tag)
            SERVER_TAG="$2"
            shift 2
            ;;
        *)
            echo "错误: 未知参数 $1"
            show_help
            exit 1
            ;;
    esac
done

# 获取服务器 IP 和位置信息
get_server_location() {
    # 获取公网 IP
    PUBLIC_IP=$(curl -s https://api.ipify.org)
    
    # 通过 IP-API 获取位置信息
    LOCATION_INFO=$(curl -s "http://ip-api.com/json/$PUBLIC_IP")
    
    # 提取国家和城市信息
    COUNTRY=$(echo $LOCATION_INFO | grep -o '"country":"[^"]*' | cut -d'"' -f4)
    CITY=$(echo $LOCATION_INFO | grep -o '"city":"[^"]*' | cut -d'"' -f4)
    
    # 组合位置标签
    if [ ! -z "$CITY" ] && [ ! -z "$COUNTRY" ]; then
        echo "$COUNTRY-$CITY"
    elif [ ! -z "$COUNTRY" ]; then
        echo "$COUNTRY"
    else
        echo "$PUBLIC_IP"
    fi
}

# 定义变量
SERVICE_NAME="loginfopush.service"
SERVICE_DIR="/etc/systemd/system"
INSTALL_DIR="/opt/loginfopush"
CONFIG_DIR="$INSTALL_DIR/config"
EXECUTABLE_URL="https://github.com/wanterfont/loginfopush/releases/download/v0.0.2/loginfopush-linux-v0.0.2-amd64"
CONFIG_URL="https://github.com/wanterfont/loginfopush/releases/download/V0.0.1/config-example-v0.0.1.json"

# 获取服务器信息（如果未通过参数指定）
DEFAULT_SERVER_NAME=$(hostname)
DEFAULT_SERVER_TAG=$(get_server_location)

# 使用参数值或提示用户输入
if [ -z "$SERVER_NAME" ]; then
    read -p "请输入服务器名称 (默认: $DEFAULT_SERVER_NAME): " SERVER_NAME
    SERVER_NAME=${SERVER_NAME:-$DEFAULT_SERVER_NAME}
fi

if [ -z "$SERVER_TAG" ]; then
    read -p "请输入服务器标签 (默认: $DEFAULT_SERVER_TAG): " SERVER_TAG
    SERVER_TAG=${SERVER_TAG:-$DEFAULT_SERVER_TAG}
fi

# 检查日志文件是否存在
if [ ! -f "/var/log/auth.log" ] && [ ! -f "/var/log/secure" ]; then
    echo "错误: 未找到 /var/log/auth.log 或 /var/log/secure 文件。"
    echo "请安装 rsyslog 以满足日志文件需求。"
    exit 1
fi

# 删除已存在的目录（如果存在）
rm -rf "$INSTALL_DIR"

# 创建所需目录
mkdir -p "$INSTALL_DIR" "$CONFIG_DIR"

# 下载文件
echo "正在下载程序文件..."
if ! curl -L "$EXECUTABLE_URL" -o "$INSTALL_DIR/loginfopush" --silent --retry 3 -f; then
    echo "错误: 程序文件下载失败"
    exit 1
fi

echo "正在下载配置文件..."
if ! curl -L "$CONFIG_URL" -o "$CONFIG_DIR/config.json" --silent --retry 3 -f; then
    echo "错误: 配置文件下载失败"
    exit 1
fi

if [ ! -s "$INSTALL_DIR/loginfopush" ] || [ ! -s "$CONFIG_DIR/config.json" ]; then
    echo "错误: 文件下载不完整"
    exit 1
fi

chmod +x "$INSTALL_DIR/loginfopush"

# 更新 server 配置
sed -i "s/\"name\": \".*\"/\"name\": \"$SERVER_NAME\"/" "$CONFIG_DIR/config.json"
sed -i "s/\"tag\": \".*\"/\"tag\": \"$SERVER_TAG\"/" "$CONFIG_DIR/config.json"

# 定义一个数组存储已启用的通知渠道
ENABLED_NOTIFIERS=()

# FCM 配置
if [ ! -z "$FCM_WEBHOOK" ] && [ ! -z "$FCM_TOKEN" ]; then
    # 如果有参数传入，直接使用参数值
    sed -i "s|\"fcm\": {.*\"enabled\": .*,|\"fcm\": {\"type\": \"fcm\", \"enabled\": true,|" "$CONFIG_DIR/config.json"
    sed -i "/\"fcm\": {/,/}/ s|\"webhook_url\": \".*\"|\"webhook_url\": \"$FCM_WEBHOOK\"|" "$CONFIG_DIR/config.json"
    sed -i "/\"fcm\": {/,/}/ s|\"device_token\": \".*\"|\"device_token\": \"$FCM_TOKEN\"|" "$CONFIG_DIR/config.json"
    ENABLED_NOTIFIERS+=("fcm")
else
    # 如果没有参数，才询问用户
    read -p "是否启用 FCM 通知? (y/n): " ENABLE_FCM
    if [[ $ENABLE_FCM == "y" ]]; then
        read -p "请输入 FCM webhook_url: " FCM_WEBHOOK
        read -p "请输入 FCM device_token: " FCM_TOKEN
        sed -i "s|\"fcm\": {.*\"enabled\": .*,|\"fcm\": {\"type\": \"fcm\", \"enabled\": true,|" "$CONFIG_DIR/config.json"
        sed -i "/\"fcm\": {/,/}/ s|\"webhook_url\": \".*\"|\"webhook_url\": \"$FCM_WEBHOOK\"|" "$CONFIG_DIR/config.json"
        sed -i "/\"fcm\": {/,/}/ s|\"device_token\": \".*\"|\"device_token\": \"$FCM_TOKEN\"|" "$CONFIG_DIR/config.json"
        ENABLED_NOTIFIERS+=("fcm")
    else
        sed -i "s|\"fcm\": {.*\"enabled\": .*,|\"fcm\": {\"type\": \"fcm\", \"enabled\": false,|" "$CONFIG_DIR/config.json"
    fi
fi

# Telegram 配置
if [ ! -z "$TG_WEBHOOK" ] && [ ! -z "$TG_CHATID" ]; then
    # 如果有参数传入，直接使用参数值
    sed -i "s|\"telegram\": {.*\"enabled\": .*,|\"telegram\": {\"type\": \"telegram\", \"enabled\": true,|" "$CONFIG_DIR/config.json"
    sed -i "/\"telegram\": {/,/}/ s|\"webhook_url\": \".*\"|\"webhook_url\": \"$TG_WEBHOOK\"|" "$CONFIG_DIR/config.json"
    sed -i "/\"telegram\": {/,/}/ s|\"chat_id\": \".*\"|\"chat_id\": \"$TG_CHATID\"|" "$CONFIG_DIR/config.json"
    ENABLED_NOTIFIERS+=("telegram")
else
    # 如果没有参数，才询问用户
    read -p "是否启用 Telegram 通知? (y/n): " ENABLE_TG
    if [[ $ENABLE_TG == "y" ]]; then
        read -p "请输入 Telegram webhook_url: " TG_WEBHOOK
        read -p "请输入 Telegram chat_id: " TG_CHATID
        sed -i "s|\"telegram\": {.*\"enabled\": .*,|\"telegram\": {\"type\": \"telegram\", \"enabled\": true,|" "$CONFIG_DIR/config.json"
        sed -i "/\"telegram\": {/,/}/ s|\"webhook_url\": \".*\"|\"webhook_url\": \"$TG_WEBHOOK\"|" "$CONFIG_DIR/config.json"
        sed -i "/\"telegram\": {/,/}/ s|\"chat_id\": \".*\"|\"chat_id\": \"$TG_CHATID\"|" "$CONFIG_DIR/config.json"
        ENABLED_NOTIFIERS+=("telegram")
    else
        sed -i "s|\"telegram\": {.*\"enabled\": .*,|\"telegram\": {\"type\": \"telegram\", \"enabled\": false,|" "$CONFIG_DIR/config.json"
    fi
fi

# Bark 配置
if [ ! -z "$BARK_WEBHOOK" ] && [ ! -z "$BARK_TOKEN" ]; then
    # 如果有参数传入，直接使用参数值
    sed -i "s|\"bark\": {.*\"enabled\": .*,|\"bark\": {\"type\": \"bark\", \"enabled\": true,|" "$CONFIG_DIR/config.json"
    sed -i "/\"bark\": {/,/}/ s|\"webhook_url\": \".*\"|\"webhook_url\": \"$BARK_WEBHOOK\"|" "$CONFIG_DIR/config.json"
    sed -i "/\"bark\": {/,/}/ s|\"device_token\": \".*\"|\"device_token\": \"$BARK_TOKEN\"|" "$CONFIG_DIR/config.json"
    ENABLED_NOTIFIERS+=("bark")
else
    # 如果没有参数，才询问用户
    read -p "是否启用 Bark 通知? (y/n): " ENABLE_BARK
    if [[ $ENABLE_BARK == "y" ]]; then
        read -p "请输入 Bark webhook_url: " BARK_WEBHOOK
        read -p "请输入 Bark device_token: " BARK_TOKEN
        sed -i "s|\"bark\": {.*\"enabled\": .*,|\"bark\": {\"type\": \"bark\", \"enabled\": true,|" "$CONFIG_DIR/config.json"
        sed -i "/\"bark\": {/,/}/ s|\"webhook_url\": \".*\"|\"webhook_url\": \"$BARK_WEBHOOK\"|" "$CONFIG_DIR/config.json"
        sed -i "/\"bark\": {/,/}/ s|\"device_token\": \".*\"|\"device_token\": \"$BARK_TOKEN\"|" "$CONFIG_DIR/config.json"
        ENABLED_NOTIFIERS+=("bark")
    else
        sed -i "s|\"bark\": {.*\"enabled\": .*,|\"bark\": {\"type\": \"bark\", \"enabled\": false,|" "$CONFIG_DIR/config.json"
    fi
fi

# WxPusher 配置
if [ ! -z "$WXPUSHER_TOKEN" ] && [ ! -z "$WXPUSHER_UIDS" ]; then
    # 如果有参数传入，直接使用参数值
    sed -i "s|\"wxpusher\": {.*\"enabled\": .*,|\"wxpusher\": {\"type\": \"wxpusher\", \"enabled\": true,|" "$CONFIG_DIR/config.json"
    sed -i "/\"wxpusher\": {/,/}/ s|\"app_token\": \".*\"|\"app_token\": \"$WXPUSHER_TOKEN\"|" "$CONFIG_DIR/config.json"
    sed -i "/\"wxpusher\": {/,/}/ s|\"uids\": \[.*\]|\"uids\": [$WXPUSHER_UIDS]|" "$CONFIG_DIR/config.json"
    ENABLED_NOTIFIERS+=("wxpusher")
else
    # 如果没有参数，才询问用户
    read -p "是否启用 WxPusher 通知? (y/n): " ENABLE_WXPUSHER
    if [[ $ENABLE_WXPUSHER == "y" ]]; then
        read -p "请输入 WxPusher app_token: " WXPUSHER_TOKEN
        
        # 初始化空数组
        WXPUSHER_UIDS=()
        
        while true; do
            read -p "请输入 WxPusher uid (输入空行结束): " WXPUSHER_UID
            if [ -z "$WXPUSHER_UID" ]; then
                break
            fi
            WXPUSHER_UIDS+=("\"$WXPUSHER_UID\"")
        done
        
        # 检查是否至少输入了一个 UID
        if [ ${#WXPUSHER_UIDS[@]} -eq 0 ]; then
            echo "错误: 至少需要输入一个 WxPusher UID"
            exit 1
        fi
        
        # 将数组转换为 JSON 格式
        UIDS_JSON=$(IFS=,; echo "[${WXPUSHER_UIDS[*]}]")
        
        sed -i "s|\"wxpusher\": {.*\"enabled\": .*,|\"wxpusher\": {\"type\": \"wxpusher\", \"enabled\": true,|" "$CONFIG_DIR/config.json"
        sed -i "/\"wxpusher\": {/,/}/ s|\"app_token\": \".*\"|\"app_token\": \"$WXPUSHER_TOKEN\"|" "$CONFIG_DIR/config.json"
        sed -i "/\"wxpusher\": {/,/}/ s|\"uids\": \[.*\]|\"uids\": $UIDS_JSON|" "$CONFIG_DIR/config.json"
        ENABLED_NOTIFIERS+=("wxpusher")
    else
        sed -i "s|\"wxpusher\": {.*\"enabled\": .*,|\"wxpusher\": {\"type\": \"wxpusher\", \"enabled\": false,|" "$CONFIG_DIR/config.json"
    fi
fi

# 配置事件类型
configure_event() {
    local event_name=$1
    local event_type=$2
    local default_enabled=$3

    read -p "是否启用${event_name}通知? (y/n, 默认: y): " ENABLE_EVENT
    ENABLE_EVENT=${ENABLE_EVENT:-y}  # 如果用户直接回车，默认为 y

    if [[ $ENABLE_EVENT == "y" ]]; then
        sed -i "s|\"$event_type\": {.*\"enabled\": .*,|\"$event_type\": {\"type\": \"$event_type\", \"enabled\": true,|" "$CONFIG_DIR/config.json"

        echo "可用的通知渠道: ${ENABLED_NOTIFIERS[@]}"
        read -p "请输入要使用的通知渠道(用空格分隔): " SELECTED_NOTIFIERS

        # 验证输入的通知渠道是否有效
        VALID_NOTIFIERS=()
        for notifier in $SELECTED_NOTIFIERS; do
            if [[ " ${ENABLED_NOTIFIERS[@]} " =~ " ${notifier} " ]]; then
                VALID_NOTIFIERS+=("\"$notifier\"")
            else
                echo "警告: 忽略未启用的通知渠道 $notifier"
            fi
        done

        # 更新事件的通知渠道配置
        NOTIFIERS_JSON=$(IFS=,; echo "[${VALID_NOTIFIERS[*]}]")
        sed -i "s|\"notifiers\": \[.*\]|\"notifiers\": $NOTIFIERS_JSON|" "$CONFIG_DIR/config.json"
    else
        sed -i "s|\"$event_type\": {.*\"enabled\": .*,|\"$event_type\": {\"type\": \"$event_type\", \"enabled\": false,|" "$CONFIG_DIR/config.json"
    fi
}

# 配置各类事件
configure_event "登录成功" "login_success" "true"

# 创建 systemd 服务文件
cat <<EOF > "$SERVICE_DIR/$SERVICE_NAME"
[Unit]
Description=loginfopush Service

[Service]
ExecStart=$INSTALL_DIR/loginfopush
Restart=always
User=root
WorkingDirectory=$INSTALL_DIR

[Install]
WantedBy=multi-user.target
EOF

# 重新加载 systemd 配置并启动服务
systemctl daemon-reload
systemctl enable "$SERVICE_NAME"
systemctl start "$SERVICE_NAME"
systemctl restart "$SERVICE_NAME"

echo "loginfopush 服务已安装并启动。"

# 生成配置命令
echo -e "\n生成用于其他机器安装的配置命令："
echo "复制以下命令可在其他机器上快速安装（记得替换 --server-name 和 --server-tag 的值）："

INSTALL_CMD="curl -fsSL https://wanterfont.github.io/loginfopush/install.sh -o install.sh && bash install.sh"

# 提取配置块的函数
extract_config() {
    local service=$1
    local config_block=$(sed -n "/\"$service\": {/,/}/p" "$CONFIG_DIR/config.json")
    echo "$config_block"
}

# 只提取已启用的通知渠道配置
for notifier in "${ENABLED_NOTIFIERS[@]}"; do
    case $notifier in
        "fcm")
            FCM_CONFIG=$(extract_config "fcm")
            FCM_WEBHOOK=$(echo "$FCM_CONFIG" | grep -o '"webhook_url": *"[^"]*"' | cut -d'"' -f4)
            FCM_TOKEN=$(echo "$FCM_CONFIG" | grep -o '"device_token": *"[^"]*"' | cut -d'"' -f4)
            INSTALL_CMD="$INSTALL_CMD --fcm-webhook '$FCM_WEBHOOK' --fcm-token '$FCM_TOKEN'"
            ;;
        "telegram")
            TG_CONFIG=$(extract_config "telegram")
            TG_WEBHOOK=$(echo "$TG_CONFIG" | grep -o '"webhook_url": *"[^"]*"' | cut -d'"' -f4)
            TG_CHATID=$(echo "$TG_CONFIG" | grep -o '"chat_id": *"[^"]*"' | cut -d'"' -f4)
            INSTALL_CMD="$INSTALL_CMD --tg-webhook '$TG_WEBHOOK' --tg-chatid '$TG_CHATID'"
            ;;
        "bark")
            BARK_CONFIG=$(extract_config "bark")
            BARK_WEBHOOK=$(echo "$BARK_CONFIG" | grep -o '"webhook_url": *"[^"]*"' | cut -d'"' -f4)
            BARK_TOKEN=$(echo "$BARK_CONFIG" | grep -o '"device_token": *"[^"]*"' | cut -d'"' -f4)
            INSTALL_CMD="$INSTALL_CMD --bark-webhook '$BARK_WEBHOOK' --bark-token '$BARK_TOKEN'"
            ;;
        "wxpusher")
            WXPUSHER_CONFIG=$(extract_config "wxpusher")
            WXPUSHER_TOKEN=$(echo "$WXPUSHER_CONFIG" | grep -o '"app_token": *"[^"]*"' | cut -d'"' -f4)
            WXPUSHER_UIDS=$(echo "$WXPUSHER_CONFIG" | grep -o '"uids": *\[[^]]*\]' | grep -o '"UID[^"]*"' | tr '\n' ',' | sed 's/,$//')
            INSTALL_CMD="$INSTALL_CMD --wxpusher-token '$WXPUSHER_TOKEN' --wxpusher-uids '$WXPUSHER_UIDS'"
            ;;
    esac
done

# 输出命令前先检查是否有推送渠道配置
if [ ${#ENABLED_NOTIFIERS[@]} -eq 0 ]; then
    echo "警告：未启用任何推送渠道！"
    exit 1
fi

echo "$INSTALL_CMD"
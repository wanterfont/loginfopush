#!/bin/bash

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
EXECUTABLE_URL="https://loginfopus.xx/loginfopush"
CONFIG_URL="https://loginfopus.xx/config.json"

# 获取服务器信息
DEFAULT_SERVER_NAME=$(hostname)
DEFAULT_SERVER_TAG=$(get_server_location)

# 提示用户输入 server 配置信息
read -p "请输入服务器名称 (默认: $DEFAULT_SERVER_NAME): " SERVER_NAME
SERVER_NAME=${SERVER_NAME:-$DEFAULT_SERVER_NAME}

read -p "请输入服务器标签 (默认: $DEFAULT_SERVER_TAG): " SERVER_TAG
SERVER_TAG=${SERVER_TAG:-$DEFAULT_SERVER_TAG}

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
declare -a ENABLED_NOTIFIERS=()

# FCM 配置
read -p "是否启用 FCM 通知? (y/n): " ENABLE_FCM
if [[ $ENABLE_FCM == "y" ]]; then
    read -p "请输入 FCM webhook_url: " FCM_WEBHOOK
    read -p "请输入 FCM device_token: " FCM_TOKEN
    sed -i "s|\"fcm\": {.*\"enabled\": .*,|\"fcm\": {\"type\": \"fcm\", \"enabled\": true,|" "$CONFIG_DIR/config.json"
    sed -i "s|\"webhook_url\": \".*\"|\"webhook_url\": \"$FCM_WEBHOOK\"|" "$CONFIG_DIR/config.json"
    sed -i "s|\"device_token\": \".*\"|\"device_token\": \"$FCM_TOKEN\"|" "$CONFIG_DIR/config.json"
    ENABLED_NOTIFIERS+=("fcm")
else
    sed -i "s|\"fcm\": {.*\"enabled\": .*,|\"fcm\": {\"type\": \"fcm\", \"enabled\": false,|" "$CONFIG_DIR/config.json"
fi

# Telegram 配置
read -p "是否启用 Telegram 通知? (y/n): " ENABLE_TG
if [[ $ENABLE_TG == "y" ]]; then
    read -p "请输入 Telegram webhook_url: " TG_WEBHOOK
    read -p "请输入 Telegram chat_id: " TG_CHATID
    sed -i "s|\"telegram\": {.*\"enabled\": .*,|\"telegram\": {\"type\": \"telegram\", \"enabled\": true,|" "$CONFIG_DIR/config.json"
    sed -i "s|\"webhook_url\": \".*\"|\"webhook_url\": \"$TG_WEBHOOK\"|" "$CONFIG_DIR/config.json"
    sed -i "s|\"chat_id\": \".*\"|\"chat_id\": \"$TG_CHATID\"|" "$CONFIG_DIR/config.json"
    ENABLED_NOTIFIERS+=("telegram")
else
    sed -i "s|\"telegram\": {.*\"enabled\": .*,|\"telegram\": {\"type\": \"telegram\", \"enabled\": false,|" "$CONFIG_DIR/config.json"
fi

# Bark 配置
read -p "是否启用 Bark 通知? (y/n): " ENABLE_BARK
if [[ $ENABLE_BARK == "y" ]]; then
    read -p "请输入 Bark webhook_url: " BARK_WEBHOOK
    read -p "请输入 Bark device_token: " BARK_TOKEN
    sed -i "s|\"bark\": {.*\"enabled\": .*,|\"bark\": {\"type\": \"bark\", \"enabled\": true,|" "$CONFIG_DIR/config.json"
    sed -i "s|\"webhook_url\": \".*\"|\"webhook_url\": \"$BARK_WEBHOOK\"|" "$CONFIG_DIR/config.json"
    sed -i "s|\"device_token\": \".*\"|\"device_token\": \"$BARK_TOKEN\"|" "$CONFIG_DIR/config.json"
    ENABLED_NOTIFIERS+=("bark")
else
    sed -i "s|\"bark\": {.*\"enabled\": .*,|\"bark\": {\"type\": \"bark\", \"enabled\": false,|" "$CONFIG_DIR/config.json"
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

echo "loginfopush 服务已安装并启动。"
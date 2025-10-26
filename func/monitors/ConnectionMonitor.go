package monitors

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"loginfopush/config"
)

// CheckConnections 检查 TCP 和 UDP 连接数
func CheckConnections(cfg config.ConnectionMonitorConfig) ([]string, error) {
	if runtime.GOOS != "linux" {
		return nil, nil // Not an error, just don't run on non-linux systems
	}
	if !cfg.Enabled {
		return nil, nil
	}

	var messages []string

	tcpCount, err := getTCPConnectionCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get TCP connection count: %w", err)
	}

	udpCount, err := getUDPConnectionCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get UDP connection count: %w", err)
	}

	totalCount := tcpCount + udpCount

	if cfg.TCPThreshold > 0 && tcpCount > cfg.TCPThreshold {
		messages = append(messages, fmt.Sprintf("TCP connection count (%d) exceeds threshold (%d)", tcpCount, cfg.TCPThreshold))
	}

	if cfg.UDPThreshold > 0 && udpCount > cfg.UDPThreshold {
		messages = append(messages, fmt.Sprintf("UDP connection count (%d) exceeds threshold (%d)", udpCount, cfg.UDPThreshold))
	}

	if cfg.TotalThreshold > 0 && totalCount > cfg.TotalThreshold {
		messages = append(messages, fmt.Sprintf("Total connection count (%d) exceeds threshold (%d)", totalCount, cfg.TotalThreshold))
	}

	return messages, nil
}

func getTCPConnectionCount() (int, error) {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("sh", "-c", "netstat -ant | grep '^tcp' | wc -l")
		return executeAndParse(cmd)
	default:
		return 0, fmt.Errorf("unsupported operating system for connection monitoring: %s", runtime.GOOS)
	}
}

func getUDPConnectionCount() (int, error) {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("sh", "-c", "netstat -anu | grep '^udp' | wc -l")
		return executeAndParse(cmd)
	default:
		return 0, fmt.Errorf("unsupported operating system for connection monitoring: %s", runtime.GOOS)
	}
}

func executeAndParse(cmd *exec.Cmd) (int, error) {
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return 0, err
	}

	// Trim space and newline characters, then parse to int
	countStr := strings.TrimSpace(out.String())
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse count: %s, error: %w", countStr, err)
	}

	return count, nil
}

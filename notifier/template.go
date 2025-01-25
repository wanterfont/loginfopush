package notifier

import (
	"bytes"
	"loginfopush/config"
	"text/template"
)

// TemplateData 模板数据结构
type TemplateData struct {
	Server   config.ServerConfig    // 服务器信息
	IP       string                 // IP 地址
	Location string                 // 位置
	Time     string                 // 时间
	Details  string                 // 详细信息
	Raw      string                 // 原始日志
	Extra    map[string]interface{} // 额外数据
}

// RenderTemplate 渲染模板
func RenderTemplate(tmpl string, data TemplateData) (string, error) {
	t, err := template.New("message").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

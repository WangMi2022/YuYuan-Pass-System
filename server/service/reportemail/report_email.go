package reportemail

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"net/mail"
	"strings"

	serverConfig "github.com/WangMi2022/mit-assets-admin/server/config"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	emailUtils "github.com/WangMi2022/mit-assets-admin/server/plugin/email/utils"
	"github.com/yuin/goldmark"
)

// Message is trusted report content assembled by a server-side report
// provider. The HTTP interface never accepts these fields from clients.
type Message struct {
	Subject  string
	Title    string
	Subtitle string
	Body     string
}

var sendSMTP = emailUtils.SendWithConfig

// SendToSystemInbox sends a report to the recipients configured in Basic
// Settings. Configuration and recipients are read at send time so changes do
// not require a process restart.
func SendToSystemInbox(message Message) error {
	configuration := currentEmailConfig()
	recipients, err := normalizeRecipients(configuration.To)
	if err != nil {
		return fmt.Errorf("系统收件邮箱配置无效: %w", err)
	}
	return send(configuration, recipients, message)
}

// SendToMailbox sends a report to a server-selected mailbox, such as the
// authenticated user's bound email address for a subscription.
func SendToMailbox(rawRecipient string, message Message) error {
	configuration := currentEmailConfig()
	recipients, err := normalizeRecipients(rawRecipient)
	if err != nil {
		return fmt.Errorf("收件邮箱配置无效: %w", err)
	}
	return send(configuration, recipients, message)
}

func send(configuration serverConfig.Email, recipients []string, message Message) error {
	if !configurationReady(configuration) {
		return errors.New("邮件服务未配置")
	}
	message.Subject = strings.TrimSpace(message.Subject)
	message.Title = strings.TrimSpace(message.Title)
	if message.Subject == "" || message.Title == "" {
		return errors.New("报告邮件主题不能为空")
	}
	return sendSMTP(configuration, recipients, message.Subject, renderHTML(message))
}

func currentEmailConfig() serverConfig.Email {
	global.GVA_CONFIG_LOCK.Lock()
	defer global.GVA_CONFIG_LOCK.Unlock()
	return global.GVA_CONFIG.Email
}

func configurationReady(configuration serverConfig.Email) bool {
	return strings.TrimSpace(configuration.From) != "" &&
		strings.TrimSpace(configuration.Host) != "" &&
		strings.TrimSpace(configuration.Secret) != "" &&
		configuration.Port > 0 && configuration.Port <= 65535
}

func normalizeRecipients(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("请先在基础设置中配置收件邮箱")
	}
	addresses, err := mail.ParseAddressList(raw)
	if err != nil {
		return nil, errors.New("请输入有效邮箱地址")
	}
	recipients := make([]string, 0, len(addresses))
	seen := make(map[string]struct{}, len(addresses))
	for _, address := range addresses {
		normalized := strings.ToLower(strings.TrimSpace(address.Address))
		if normalized == "" {
			return nil, errors.New("请输入有效邮箱地址")
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		recipients = append(recipients, normalized)
	}
	if len(recipients) == 0 {
		return nil, errors.New("请先在基础设置中配置收件邮箱")
	}
	return recipients, nil
}

func renderHTML(message Message) string {
	body := strings.TrimSpace(message.Body)
	if body == "" {
		body = "报告暂无正文。"
	}
	body = stripDuplicateMarkdownTitle(body, message.Title)
	subtitle := ""
	if value := strings.TrimSpace(message.Subtitle); value != "" {
		subtitle = fmt.Sprintf(`<p style="margin:0 0 20px;color:#64748b">%s</p>`, html.EscapeString(value))
	}
	return fmt.Sprintf(
		`<div style="font-family:Arial,'Microsoft YaHei',sans-serif;color:#1f2937;line-height:1.7;max-width:760px;margin:0 auto"><h1 style="font-size:22px;margin:0 0 8px">%s</h1>%s<div style="padding:18px;border:1px solid #e2e8f0;border-radius:8px;background:#f8fafc;overflow-wrap:anywhere">%s</div><p style="margin:18px 0 0;color:#94a3b8;font-size:12px">由系统自动生成并发送</p></div>`,
		html.EscapeString(message.Title),
		subtitle,
		renderMarkdown(body),
	)
}

func stripDuplicateMarkdownTitle(body, title string) string {
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	if len(lines) == 0 {
		return body
	}
	first := strings.TrimSpace(lines[0])
	heading := strings.TrimSpace(strings.TrimPrefix(first, "# "))
	if !strings.HasPrefix(first, "# ") || (heading != strings.TrimSpace(title) && heading != "今日"+strings.TrimSpace(title)) {
		return body
	}
	lines = lines[1:]
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	return strings.Join(lines, "\n")
}

func renderMarkdown(markdown string) string {
	var output bytes.Buffer
	// Escape raw HTML before parsing so model-produced tags remain visible text
	// while Markdown structure can still be rendered safely.
	safeMarkdown := html.EscapeString(markdown)
	if err := goldmark.Convert([]byte(safeMarkdown), &output); err != nil {
		return `<p style="margin:0">` + safeMarkdown + `</p>`
	}
	return styleMarkdownHTML(output.String())
}

func styleMarkdownHTML(value string) string {
	return strings.NewReplacer(
		"<h1>", `<h2 style="font-size:20px;line-height:1.4;margin:0 0 14px;color:#111827">`,
		"</h1>", "</h2>",
		"<h2>", `<h2 style="font-size:17px;line-height:1.5;margin:20px 0 8px;color:#111827">`,
		"<h3>", `<h3 style="font-size:15px;line-height:1.5;margin:16px 0 6px;color:#1f2937">`,
		"<p>", `<p style="margin:0 0 12px">`,
		"<ul>", `<ul style="margin:6px 0 14px;padding-left:22px">`,
		"<ol>", `<ol style="margin:6px 0 14px;padding-left:22px">`,
		"<li>", `<li style="margin:4px 0">`,
		"<blockquote>", `<blockquote style="margin:12px 0;padding:8px 12px;border-left:3px solid #94a3b8;color:#475569;background:#f1f5f9">`,
		"<code>", `<code style="font-family:Consolas,monospace;padding:1px 4px;background:#e2e8f0;border-radius:3px">`,
		"<strong>", `<strong style="color:#111827;font-weight:700">`,
		"<a ", `<a style="color:#2563eb;text-decoration:underline" `,
		"<hr>\n", `<hr style="border:0;border-top:1px solid #cbd5e1;margin:18px 0">`,
	).Replace(value)
}

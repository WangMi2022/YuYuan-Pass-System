package reportemail

import (
	"errors"
	"fmt"
	"html"
	"net/mail"
	"strings"

	serverConfig "github.com/WangMi2022/mit-assets-admin/server/config"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	emailUtils "github.com/WangMi2022/mit-assets-admin/server/plugin/email/utils"
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
	subtitle := ""
	if value := strings.TrimSpace(message.Subtitle); value != "" {
		subtitle = fmt.Sprintf(`<p style="margin:0 0 20px;color:#64748b">%s</p>`, html.EscapeString(value))
	}
	return fmt.Sprintf(
		`<div style="font-family:Arial,'Microsoft YaHei',sans-serif;color:#1f2937;line-height:1.7;max-width:760px;margin:0 auto"><h1 style="font-size:22px;margin:0 0 8px">%s</h1>%s<div style="padding:18px;border:1px solid #e2e8f0;border-radius:10px;background:#f8fafc;white-space:pre-wrap">%s</div><p style="margin:18px 0 0;color:#94a3b8;font-size:12px">由系统自动生成并发送</p></div>`,
		html.EscapeString(message.Title),
		subtitle,
		html.EscapeString(body),
	)
}

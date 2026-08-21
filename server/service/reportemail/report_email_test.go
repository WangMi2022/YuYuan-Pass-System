package reportemail

import (
	"strings"
	"testing"

	serverConfig "github.com/WangMi2022/mit-assets-admin/server/config"
	"github.com/WangMi2022/mit-assets-admin/server/global"
)

func TestSendToSystemInboxUsesRuntimeConfigAndSanitizesContent(t *testing.T) {
	global.GVA_CONFIG_LOCK.Lock()
	previousEmail := global.GVA_CONFIG.Email
	global.GVA_CONFIG.Email = serverConfig.Email{
		To:     "Leader <Leader@Example.com>, audit@example.com, leader@example.com",
		From:   "sender@example.com",
		Host:   "smtp.example.com",
		Secret: "smtp-secret",
		Port:   465,
		IsSSL:  true,
	}
	global.GVA_CONFIG_LOCK.Unlock()
	previousSender := sendSMTP
	t.Cleanup(func() {
		global.GVA_CONFIG_LOCK.Lock()
		global.GVA_CONFIG.Email = previousEmail
		global.GVA_CONFIG_LOCK.Unlock()
		sendSMTP = previousSender
	})

	var actualRecipients []string
	var actualBody string
	sendSMTP = func(configuration serverConfig.Email, recipients []string, subject, body string) error {
		if configuration.Host != "smtp.example.com" || subject != "风险报告 - 2026-08-21" {
			t.Fatalf("unexpected email request: host=%q subject=%q", configuration.Host, subject)
		}
		actualRecipients = append([]string(nil), recipients...)
		actualBody = body
		return nil
	}

	err := SendToSystemInbox(Message{
		Subject:  "风险报告 - 2026-08-21",
		Title:    "风险报告<script>",
		Subtitle: "生成时间：2026-08-21 14:05:06",
		Body:     "开放风险 3 条\n<script>alert('x')</script>",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(actualRecipients) != 2 || actualRecipients[0] != "leader@example.com" || actualRecipients[1] != "audit@example.com" {
		t.Fatalf("unexpected recipients: %#v", actualRecipients)
	}
	if strings.Contains(actualBody, "<script>") || !strings.Contains(actualBody, "&lt;script&gt;") || !strings.Contains(actualBody, "2026-08-21 14:05:06") {
		t.Fatalf("report email body was not rendered safely: %q", actualBody)
	}
}

func TestSendToSystemInboxRejectsMissingRecipients(t *testing.T) {
	global.GVA_CONFIG_LOCK.Lock()
	previousEmail := global.GVA_CONFIG.Email
	global.GVA_CONFIG.Email = serverConfig.Email{
		From: "sender@example.com", Host: "smtp.example.com", Secret: "smtp-secret", Port: 465,
	}
	global.GVA_CONFIG_LOCK.Unlock()
	t.Cleanup(func() {
		global.GVA_CONFIG_LOCK.Lock()
		global.GVA_CONFIG.Email = previousEmail
		global.GVA_CONFIG_LOCK.Unlock()
	})

	err := SendToSystemInbox(Message{Subject: "日报", Title: "日报", Body: "正文"})
	if err == nil || !strings.Contains(err.Error(), "基础设置") {
		t.Fatalf("expected a configured-recipient error, got %v", err)
	}
}

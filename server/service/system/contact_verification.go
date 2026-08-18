package system

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/config"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	systemRes "github.com/WangMi2022/mit-assets-admin/server/model/system/response"
	emailUtils "github.com/WangMi2022/mit-assets-admin/server/plugin/email/utils"
)

const (
	contactChannelPhone = "phone"
	contactChannelEmail = "email"

	contactVerificationCodeTTL     = 5 * time.Minute
	contactVerificationResendDelay = 60 * time.Second
	contactVerificationMaxAttempts = 5
)

var (
	phonePattern             = regexp.MustCompile(`^1[3-9]\d{9}$`)
	verificationCodePattern  = regexp.MustCompile(`^\d{6}$`)
	verificationChannelLocks sync.Map
	verificationHTTPClient   = &http.Client{Timeout: 8 * time.Second}
)

type contactVerificationRecord struct {
	Target       string
	CodeHash     [sha256.Size]byte
	ExpiresAt    time.Time
	AttemptsLeft int
}

type smsVerificationPayload struct {
	Phone            string `json:"phone"`
	Code             string `json:"code"`
	SignName         string `json:"signName"`
	TemplateID       string `json:"templateId"`
	Purpose          string `json:"purpose"`
	ExpiresInSeconds int    `json:"expiresInSeconds"`
}

func (userService *UserService) ContactVerificationCapabilities() systemRes.ContactVerificationCapabilities {
	verification, smtp := currentContactVerificationConfig()
	phoneConfigured := verification.SMS.Ready()
	emailConfigured := verification.Email.Ready(smtp)
	return systemRes.ContactVerificationCapabilities{
		Phone: contactVerificationCapability(
			phoneConfigured,
			verification.Enabled,
			verification.SMS.Enabled,
			"短信服务尚未完成配置",
			"短信验证码尚未开启",
			"短信验证码已启用",
		),
		Email: contactVerificationCapability(
			emailConfigured,
			verification.Enabled,
			verification.Email.Enabled,
			"SMTP 服务尚未完成配置",
			"邮件验证码尚未开启",
			"邮件验证码已启用",
		),
	}
}

func (userService *UserService) SendContactVerificationCode(
	ctx context.Context,
	userID uint,
	channel string,
	target string,
) error {
	channel, target, err := normalizeContactVerificationTarget(channel, target)
	if err != nil {
		return err
	}
	lock := contactVerificationLock(userID, channel)
	lock.Lock()
	defer lock.Unlock()

	if _, exists := global.BlackCache.Get(contactVerificationCooldownKey(userID, channel)); exists {
		return errors.New("验证码发送过于频繁，请稍后重试")
	}

	verification, smtp := currentContactVerificationConfig()
	if !verification.Enabled {
		return errors.New("联系方式验证码总开关尚未开启")
	}
	code, err := generateContactVerificationCode()
	if err != nil {
		return errors.New("验证码生成失败，请稍后重试")
	}
	switch channel {
	case contactChannelPhone:
		if !verification.SMS.Enabled || !verification.SMS.Ready() {
			return errors.New("短信验证码服务尚未配置并开启")
		}
		err = sendSMSVerificationCode(ctx, verification.SMS, target, code)
	case contactChannelEmail:
		if !verification.Email.Enabled || !verification.Email.Ready(smtp) {
			return errors.New("邮件验证码服务尚未配置并开启")
		}
		err = sendEmailVerificationCode(verification.Email, smtp, target, code)
	}
	if err != nil {
		return err
	}

	global.BlackCache.Set(contactVerificationRecordKey(userID, channel), contactVerificationRecord{
		Target:       target,
		CodeHash:     sha256.Sum256([]byte(code)),
		ExpiresAt:    time.Now().Add(contactVerificationCodeTTL),
		AttemptsLeft: contactVerificationMaxAttempts,
	}, contactVerificationCodeTTL)
	global.BlackCache.Set(
		contactVerificationCooldownKey(userID, channel),
		struct{}{},
		contactVerificationResendDelay,
	)
	return nil
}

func (userService *UserService) UpdateSelfContact(
	userID uint,
	channel string,
	target string,
	code string,
) error {
	channel, target, err := normalizeContactVerificationTarget(channel, target)
	if err != nil {
		return err
	}
	code = strings.TrimSpace(code)
	if !verificationCodePattern.MatchString(code) {
		return errors.New("请输入 6 位数字验证码")
	}
	lock := contactVerificationLock(userID, channel)
	lock.Lock()
	defer lock.Unlock()

	verification, smtp := currentContactVerificationConfig()
	if !verification.Enabled {
		return errors.New("联系方式验证码总开关尚未开启")
	}
	if channel == contactChannelPhone && (!verification.SMS.Enabled || !verification.SMS.Ready()) {
		return errors.New("短信验证码服务尚未配置并开启")
	}
	if channel == contactChannelEmail && (!verification.Email.Enabled || !verification.Email.Ready(smtp)) {
		return errors.New("邮件验证码服务尚未配置并开启")
	}

	cacheKey := contactVerificationRecordKey(userID, channel)
	value, exists := global.BlackCache.Get(cacheKey)
	if !exists {
		return errors.New("验证码已失效，请重新获取")
	}
	record, ok := value.(contactVerificationRecord)
	if !ok || time.Now().After(record.ExpiresAt) {
		global.BlackCache.Delete(cacheKey)
		return errors.New("验证码已失效，请重新获取")
	}
	if record.Target != target {
		return errors.New("验证码与当前联系方式不匹配，请重新获取")
	}

	actualHash := sha256.Sum256([]byte(code))
	if subtle.ConstantTimeCompare(record.CodeHash[:], actualHash[:]) != 1 {
		record.AttemptsLeft--
		if record.AttemptsLeft <= 0 {
			global.BlackCache.Delete(cacheKey)
			return errors.New("验证码错误次数过多，请重新获取")
		}
		remainingTTL := time.Until(record.ExpiresAt)
		if remainingTTL > 0 {
			global.BlackCache.Set(cacheKey, record, remainingTTL)
		}
		return fmt.Errorf("验证码错误，还可尝试 %d 次", record.AttemptsLeft)
	}

	column := "phone"
	if channel == contactChannelEmail {
		column = "email"
	}
	if global.GVA_DB == nil {
		return errors.New("数据库服务暂不可用")
	}
	result := global.GVA_DB.Table("sys_users").
		Where("id = ?", userID).
		Updates(map[string]interface{}{column: target, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("用户不存在")
	}
	global.BlackCache.Delete(cacheKey)
	return nil
}

func contactVerificationCapability(
	configured bool,
	masterEnabled bool,
	channelEnabled bool,
	unconfiguredReason string,
	disabledReason string,
	enabledReason string,
) systemRes.ContactVerificationChannelCapability {
	enabled := masterEnabled && channelEnabled && configured
	reason := enabledReason
	if !masterEnabled {
		reason = "联系方式验证总开关尚未开启"
	} else if !configured {
		reason = unconfiguredReason
	} else if !enabled {
		reason = disabledReason
	}
	return systemRes.ContactVerificationChannelCapability{
		Configured: configured,
		Enabled:    enabled,
		Reason:     reason,
	}
}

func currentContactVerificationConfig() (config.ContactVerification, config.Email) {
	global.GVA_CONFIG_LOCK.Lock()
	verification := global.GVA_CONFIG.ContactVerification
	smtp := global.GVA_CONFIG.Email
	global.GVA_CONFIG_LOCK.Unlock()
	verification.Normalize()
	return verification, smtp
}

func normalizeContactVerificationTarget(channel string, target string) (string, string, error) {
	channel = strings.ToLower(strings.TrimSpace(channel))
	target = strings.TrimSpace(target)
	switch channel {
	case contactChannelPhone:
		if !phonePattern.MatchString(target) {
			return "", "", errors.New("请输入有效的 11 位手机号码")
		}
	case contactChannelEmail:
		address, err := mail.ParseAddress(target)
		if err != nil || !strings.EqualFold(address.Address, target) {
			return "", "", errors.New("请输入有效的邮箱地址")
		}
		target = strings.ToLower(target)
	default:
		return "", "", errors.New("不支持的验证码类型")
	}
	return channel, target, nil
}

func generateContactVerificationCode() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}

func contactVerificationRecordKey(userID uint, channel string) string {
	return fmt.Sprintf("contact-verification:code:%d:%s", userID, channel)
}

func contactVerificationCooldownKey(userID uint, channel string) string {
	return fmt.Sprintf("contact-verification:cooldown:%d:%s", userID, channel)
}

func contactVerificationLock(userID uint, channel string) *sync.Mutex {
	key := fmt.Sprintf("%d:%s", userID, channel)
	lock, _ := verificationChannelLocks.LoadOrStore(key, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func sendEmailVerificationCode(
	verification config.ContactVerificationEmail,
	smtp config.Email,
	target string,
	code string,
) error {
	body := fmt.Sprintf(
		`<div style="font-family:Arial,'Microsoft YaHei',sans-serif;color:#1f2937;line-height:1.7"><p>您正在修改系统联系方式，本次验证码为：</p><p style="font-size:28px;font-weight:700;letter-spacing:6px">%s</p><p>验证码 5 分钟内有效，请勿转发给他人。</p></div>`,
		code,
	)
	if err := emailUtils.SendWithConfig(smtp, []string{target}, verification.Subject, body); err != nil {
		return fmt.Errorf("邮件验证码发送失败: %w", err)
	}
	return nil
}

func sendSMSVerificationCode(
	ctx context.Context,
	configuration config.ContactVerificationSMS,
	phone string,
	code string,
) error {
	payload, err := json.Marshal(smsVerificationPayload{
		Phone:            phone,
		Code:             code,
		SignName:         configuration.SignName,
		TemplateID:       configuration.TemplateID,
		Purpose:          "update_contact",
		ExpiresInSeconds: int(contactVerificationCodeTTL / time.Second),
	})
	if err != nil {
		return errors.New("短信验证码请求生成失败")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, configuration.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return errors.New("短信服务地址无效")
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+configuration.AccessToken)
	response, err := verificationHTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("短信验证码发送失败: %w", err)
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("短信服务返回异常状态: %d", response.StatusCode)
	}
	return nil
}

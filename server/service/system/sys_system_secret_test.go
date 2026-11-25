package system

import (
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/config"
)

func TestRedactSystemConfigSecrets(t *testing.T) {
	original := systemSecretFixture()
	redacted, configured := redactSystemConfigSecrets(original)

	if len(configured) != len(systemSecretAccessors) {
		t.Fatalf("configured secret count = %d, want %d", len(configured), len(systemSecretAccessors))
	}
	for path := range systemSecretAccessors {
		if !configured[path] {
			t.Errorf("secret %q was not marked configured", path)
		}
		value, ok := revealSystemConfigSecret(redacted, path)
		if !ok {
			t.Errorf("redacted secret %q is not in the whitelist", path)
		} else if value != "" {
			t.Errorf("redacted secret %q leaked value %q", path, value)
		}
	}
	if !redacted.AI.Invoice.Baidu.APIKeyConfigured ||
		!redacted.AI.Invoice.Verification.SecretKeyConfigured ||
		!redacted.AI.Invoice.Multimodal.APIKeyConfigured {
		t.Fatal("invoice credential configured flags were not preserved")
	}
}

func TestMergeSystemConfigSecretsPreservesBlankAndAcceptsReplacement(t *testing.T) {
	current := systemSecretFixture()
	incoming := config.Server{}
	incoming.Qiniu.AccessKey = "replacement-access"
	incoming.AI.Invoice = current.AI.Invoice.Redacted()

	mergeSystemConfigSecrets(&incoming, current)

	if incoming.Qiniu.AccessKey != "replacement-access" {
		t.Fatalf("explicit replacement was lost: %q", incoming.Qiniu.AccessKey)
	}
	if incoming.Qiniu.SecretKey != current.Qiniu.SecretKey {
		t.Fatalf("blank secret was not preserved: %q", incoming.Qiniu.SecretKey)
	}
	if incoming.JWT.SigningKey != current.JWT.SigningKey {
		t.Fatalf("blank JWT signing key was not preserved: %q", incoming.JWT.SigningKey)
	}

	incoming.AI.Invoice.Baidu.APIKey = ""
	mergeSystemConfigSecrets(&incoming, current)
	if incoming.AI.Invoice.Baidu.APIKey != "" {
		t.Fatalf("cleared invoice key was restored: %q", incoming.AI.Invoice.Baidu.APIKey)
	}
}

func TestRevealSystemConfigSecretUsesWhitelist(t *testing.T) {
	configuration := systemSecretFixture()

	value, ok := revealSystemConfigSecret(configuration, "qiniu.secret-key")
	if !ok || value != "qiniu-secret" {
		t.Fatalf("revealed value = %q, ok = %v", value, ok)
	}
	if _, ok := revealSystemConfigSecret(configuration, "mysql.password"); ok {
		t.Fatal("non-whitelisted field was revealed")
	}
	if _, ok := revealSystemConfigSecret(configuration, "qiniu.bucket"); ok {
		t.Fatal("non-secret field was revealed")
	}
}

func systemSecretFixture() config.Server {
	return config.Server{
		JWT:   config.JWT{SigningKey: "jwt-secret"},
		Email: config.Email{Secret: "email-secret"},
		ContactVerification: config.ContactVerification{
			SMS: config.ContactVerificationSMS{AccessToken: "sms-access-token"},
		},
		Qiniu: config.Qiniu{AccessKey: "qiniu-access", SecretKey: "qiniu-secret"},
		TencentCOS: config.TencentCOS{
			SecretID: "tencent-id", SecretKey: "tencent-secret",
		},
		AliyunOSS: config.AliyunOSS{
			AccessKeyId: "aliyun-id", AccessKeySecret: "aliyun-secret",
		},
		HuaWeiObs: config.HuaWeiObs{AccessKey: "huawei-access", SecretKey: "huawei-secret"},
		AwsS3:     config.AwsS3{SecretID: "aws-id", SecretKey: "aws-secret"},
		CloudflareR2: config.CloudflareR2{
			AccessKeyID: "r2-id", SecretAccessKey: "r2-secret",
		},
		Minio: config.Minio{AccessKeyId: "minio-id", AccessKeySecret: "minio-secret"},
		AI: config.AI{
			OpenAICompatible: config.AIProvider{APIKey: "openai-api"},
			Anthropic:        config.AIProvider{APIKey: "anthropic-api"},
			Invoice: config.InvoiceRecognition{
				Baidu:        config.InvoiceBaiduProvider{APIKey: "baidu-api", SecretKey: "baidu-secret"},
				PublicOCR:    config.InvoicePublicOCR{APIKey: "public-ocr-api"},
				Verification: config.InvoiceVerificationProvider{APIKey: "verify-api", SecretKey: "verify-secret"},
				Multimodal:   config.InvoiceMultimodalProvider{APIKey: "multimodal-api"},
			},
		},
	}
}

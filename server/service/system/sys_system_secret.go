package system

import (
	"errors"
	"strings"

	"github.com/WangMi2022/mit-assets-admin/server/config"
)

var errUnknownSystemSecret = errors.New("不支持读取该配置字段")

type systemSecretAccessor struct {
	get func(config.Server) string
	set func(*config.Server, string)
}

var systemSecretAccessors = map[string]systemSecretAccessor{
	"jwt.signing-key": {
		get: func(configuration config.Server) string { return configuration.JWT.SigningKey },
		set: func(configuration *config.Server, value string) { configuration.JWT.SigningKey = value },
	},
	"email.secret": {
		get: func(configuration config.Server) string { return configuration.Email.Secret },
		set: func(configuration *config.Server, value string) { configuration.Email.Secret = value },
	},
	"contact-verification.sms.access-token": {
		get: func(configuration config.Server) string {
			return configuration.ContactVerification.SMS.AccessToken
		},
		set: func(configuration *config.Server, value string) {
			configuration.ContactVerification.SMS.AccessToken = value
		},
	},
	"qiniu.access-key": {
		get: func(configuration config.Server) string { return configuration.Qiniu.AccessKey },
		set: func(configuration *config.Server, value string) { configuration.Qiniu.AccessKey = value },
	},
	"qiniu.secret-key": {
		get: func(configuration config.Server) string { return configuration.Qiniu.SecretKey },
		set: func(configuration *config.Server, value string) { configuration.Qiniu.SecretKey = value },
	},
	"tencent-cos.secret-id": {
		get: func(configuration config.Server) string { return configuration.TencentCOS.SecretID },
		set: func(configuration *config.Server, value string) { configuration.TencentCOS.SecretID = value },
	},
	"tencent-cos.secret-key": {
		get: func(configuration config.Server) string { return configuration.TencentCOS.SecretKey },
		set: func(configuration *config.Server, value string) { configuration.TencentCOS.SecretKey = value },
	},
	"aliyun-oss.access-key-id": {
		get: func(configuration config.Server) string { return configuration.AliyunOSS.AccessKeyId },
		set: func(configuration *config.Server, value string) { configuration.AliyunOSS.AccessKeyId = value },
	},
	"aliyun-oss.access-key-secret": {
		get: func(configuration config.Server) string { return configuration.AliyunOSS.AccessKeySecret },
		set: func(configuration *config.Server, value string) { configuration.AliyunOSS.AccessKeySecret = value },
	},
	"hua-wei-obs.access-key": {
		get: func(configuration config.Server) string { return configuration.HuaWeiObs.AccessKey },
		set: func(configuration *config.Server, value string) { configuration.HuaWeiObs.AccessKey = value },
	},
	"hua-wei-obs.secret-key": {
		get: func(configuration config.Server) string { return configuration.HuaWeiObs.SecretKey },
		set: func(configuration *config.Server, value string) { configuration.HuaWeiObs.SecretKey = value },
	},
	"aws-s3.secret-id": {
		get: func(configuration config.Server) string { return configuration.AwsS3.SecretID },
		set: func(configuration *config.Server, value string) { configuration.AwsS3.SecretID = value },
	},
	"aws-s3.secret-key": {
		get: func(configuration config.Server) string { return configuration.AwsS3.SecretKey },
		set: func(configuration *config.Server, value string) { configuration.AwsS3.SecretKey = value },
	},
	"cloudflare-r2.access-key-id": {
		get: func(configuration config.Server) string { return configuration.CloudflareR2.AccessKeyID },
		set: func(configuration *config.Server, value string) { configuration.CloudflareR2.AccessKeyID = value },
	},
	"cloudflare-r2.secret-access-key": {
		get: func(configuration config.Server) string { return configuration.CloudflareR2.SecretAccessKey },
		set: func(configuration *config.Server, value string) { configuration.CloudflareR2.SecretAccessKey = value },
	},
	"minio.access-key-id": {
		get: func(configuration config.Server) string { return configuration.Minio.AccessKeyId },
		set: func(configuration *config.Server, value string) { configuration.Minio.AccessKeyId = value },
	},
	"minio.access-key-secret": {
		get: func(configuration config.Server) string { return configuration.Minio.AccessKeySecret },
		set: func(configuration *config.Server, value string) { configuration.Minio.AccessKeySecret = value },
	},
	"ai.invoice.baidu.api-key": {
		get: func(configuration config.Server) string { return configuration.AI.Invoice.Baidu.APIKey },
		set: func(configuration *config.Server, value string) {
			configuration.AI.Invoice.Baidu.APIKey = value
		},
	},
	"ai.invoice.baidu.secret-key": {
		get: func(configuration config.Server) string { return configuration.AI.Invoice.Baidu.SecretKey },
		set: func(configuration *config.Server, value string) {
			configuration.AI.Invoice.Baidu.SecretKey = value
		},
	},
	"ai.invoice.public-ocr.api-key": {
		get: func(configuration config.Server) string { return configuration.AI.Invoice.PublicOCR.APIKey },
		set: func(configuration *config.Server, value string) {
			configuration.AI.Invoice.PublicOCR.APIKey = value
		},
	},
	"ai.invoice.verification.api-key": {
		get: func(configuration config.Server) string { return configuration.AI.Invoice.Verification.APIKey },
		set: func(configuration *config.Server, value string) {
			configuration.AI.Invoice.Verification.APIKey = value
		},
	},
	"ai.invoice.verification.secret-key": {
		get: func(configuration config.Server) string {
			return configuration.AI.Invoice.Verification.SecretKey
		},
		set: func(configuration *config.Server, value string) {
			configuration.AI.Invoice.Verification.SecretKey = value
		},
	},
	"ai.invoice.multimodal.api-key": {
		get: func(configuration config.Server) string { return configuration.AI.Invoice.Multimodal.APIKey },
		set: func(configuration *config.Server, value string) {
			configuration.AI.Invoice.Multimodal.APIKey = value
		},
	},
	"ai.openai-compatible.api-key": {
		get: func(configuration config.Server) string { return configuration.AI.OpenAICompatible.APIKey },
		set: func(configuration *config.Server, value string) {
			configuration.AI.OpenAICompatible.APIKey = value
		},
	},
	"ai.anthropic.api-key": {
		get: func(configuration config.Server) string { return configuration.AI.Anthropic.APIKey },
		set: func(configuration *config.Server, value string) { configuration.AI.Anthropic.APIKey = value },
	},
}

func redactSystemConfigSecrets(configuration config.Server) (config.Server, map[string]bool) {
	configured := make(map[string]bool, len(systemSecretAccessors))
	configuration.AI.Invoice.Normalize()
	configuration.AI.Invoice = configuration.AI.Invoice.Redacted()
	configuration.AI.Normalize()
	configuration.AI = configuration.AI.Redacted()
	for path, accessor := range systemSecretAccessors {
		if strings.TrimSpace(accessor.get(configuration)) != "" {
			configured[path] = true
		}
		accessor.set(&configuration, "")
	}
	return configuration, configured
}

func mergeSystemConfigSecrets(incoming *config.Server, current config.Server) {
	for path, accessor := range systemSecretAccessors {
		// Invoice credentials have explicit clear flags and are merged by
		// InvoiceRecognition.MergeSecrets before this generic preservation pass.
		if strings.HasPrefix(path, "ai.invoice.") {
			continue
		}
		if strings.TrimSpace(accessor.get(*incoming)) == "" {
			accessor.set(incoming, accessor.get(current))
		}
	}
}

func revealSystemConfigSecret(configuration config.Server, path string) (string, bool) {
	accessor, exists := systemSecretAccessors[strings.TrimSpace(path)]
	if !exists {
		return "", false
	}
	return accessor.get(configuration), true
}

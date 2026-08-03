package system

import (
	"errors"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
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
	"invoice-recognition.baidu.api-key": {
		get: func(configuration config.Server) string { return configuration.InvoiceRecognition.Baidu.APIKey },
		set: func(configuration *config.Server, value string) {
			configuration.InvoiceRecognition.Baidu.APIKey = value
		},
	},
	"invoice-recognition.baidu.secret-key": {
		get: func(configuration config.Server) string { return configuration.InvoiceRecognition.Baidu.SecretKey },
		set: func(configuration *config.Server, value string) {
			configuration.InvoiceRecognition.Baidu.SecretKey = value
		},
	},
	"invoice-recognition.public-ocr.api-key": {
		get: func(configuration config.Server) string { return configuration.InvoiceRecognition.PublicOCR.APIKey },
		set: func(configuration *config.Server, value string) {
			configuration.InvoiceRecognition.PublicOCR.APIKey = value
		},
	},
	"invoice-recognition.verification.api-key": {
		get: func(configuration config.Server) string { return configuration.InvoiceRecognition.Verification.APIKey },
		set: func(configuration *config.Server, value string) {
			configuration.InvoiceRecognition.Verification.APIKey = value
		},
	},
	"invoice-recognition.verification.secret-key": {
		get: func(configuration config.Server) string {
			return configuration.InvoiceRecognition.Verification.SecretKey
		},
		set: func(configuration *config.Server, value string) {
			configuration.InvoiceRecognition.Verification.SecretKey = value
		},
	},
	"invoice-recognition.multimodal.api-key": {
		get: func(configuration config.Server) string { return configuration.InvoiceRecognition.Multimodal.APIKey },
		set: func(configuration *config.Server, value string) {
			configuration.InvoiceRecognition.Multimodal.APIKey = value
		},
	},
}

func redactSystemConfigSecrets(configuration config.Server) (config.Server, map[string]bool) {
	configured := make(map[string]bool, len(systemSecretAccessors))
	configuration.InvoiceRecognition.Normalize()
	configuration.InvoiceRecognition = configuration.InvoiceRecognition.Redacted()
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
		if strings.HasPrefix(path, "invoice-recognition.") {
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

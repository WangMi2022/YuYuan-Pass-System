package system

import (
	"context"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	invoiceProvider "github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/provider"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"go.uber.org/zap"
)

//@author: [piexlmax](https://github.com/piexlmax)
//@function: GetSystemConfig
//@description: 读取配置文件
//@return: conf config.Server, err error

type SystemConfigService struct{}

type InvoiceRecognitionDetections struct {
	PublicOCR    invoiceProvider.ConnectionInfo `json:"publicOCR"`
	Verification invoiceProvider.ConnectionInfo `json:"verification"`
	Multimodal   invoiceProvider.ConnectionInfo `json:"multimodal"`
}

var SystemConfigServiceApp = new(SystemConfigService)

func (systemConfigService *SystemConfigService) GetSystemConfig() (
	conf config.Server,
	configuredSecrets map[string]bool,
	err error,
) {
	conf, configuredSecrets = redactSystemConfigSecrets(global.GVA_CONFIG)
	return conf, configuredSecrets, nil
}

func (systemConfigService *SystemConfigService) GetSystemConfigSecret(path string) (string, error) {
	value, ok := revealSystemConfigSecret(global.GVA_CONFIG, path)
	if !ok {
		return "", errUnknownSystemSecret
	}
	return value, nil
}

// @description   set system config,
//@author: [piexlmax](https://github.com/piexlmax)
//@function: SetSystemConfig
//@description: 设置配置文件
//@param: system model.System
//@return: err error

func (systemConfigService *SystemConfigService) SetSystemConfig(
	ctx context.Context,
	system system.System,
	allowInvoiceRecognition bool,
) (detections InvoiceRecognitionDetections, err error) {
	system.Config.InvoiceRecognition, err = prepareInvoiceRecognitionConfig(
		ctx,
		system.Config.InvoiceRecognition,
		global.GVA_CONFIG.InvoiceRecognition,
		allowInvoiceRecognition,
	)
	if err != nil {
		return InvoiceRecognitionDetections{}, err
	}
	mergeSystemConfigSecrets(&system.Config, global.GVA_CONFIG)
	cs := utils.StructToMap(system.Config)
	for k, v := range cs {
		global.GVA_VP.Set(k, v)
	}
	if err = global.GVA_VP.WriteConfig(); err != nil {
		return InvoiceRecognitionDetections{}, err
	}
	// The invoice worker reads this section for every job, so provider changes
	// take effect immediately without restarting database connections.
	global.GVA_CONFIG.InvoiceRecognition = system.Config.InvoiceRecognition
	invoiceProvider.SetRuntimeInvoiceRecognition(system.Config.InvoiceRecognition)
	return invoiceRecognitionDetections(system.Config.InvoiceRecognition), nil
}

func invoiceRecognitionDetections(configuration config.InvoiceRecognition) InvoiceRecognitionDetections {
	return InvoiceRecognitionDetections{
		PublicOCR: invoiceProvider.ConnectionInfo{
			Provider: configuration.PublicOCR.Provider,
			Protocol: configuration.PublicOCR.Protocol,
		},
		Verification: invoiceProvider.ConnectionInfo{
			Provider: configuration.Verification.Provider,
			Protocol: configuration.Verification.Protocol,
		},
		Multimodal: invoiceProvider.ConnectionInfo{
			Provider: "multimodal",
			Protocol: configuration.Multimodal.Protocol,
		},
	}
}

func prepareInvoiceRecognitionConfig(
	ctx context.Context,
	incoming config.InvoiceRecognition,
	current config.InvoiceRecognition,
	allow bool,
) (config.InvoiceRecognition, error) {
	current.Normalize()
	prepared := incoming.MergeSecrets(current, allow)
	prepared.Normalize()
	if allow && shouldDetectPublicOCR(current.PublicOCR, prepared.PublicOCR) {
		probeConfiguration := prepared
		probeConfiguration.PublicOCR.Provider = ""
		probeConfiguration.PublicOCR.Protocol = ""
		detection, err := invoiceProvider.TestConnection(ctx, "public-ocr", probeConfiguration)
		if err != nil {
			return config.InvoiceRecognition{}, fmt.Errorf("公网 OCR 接口自动探测失败: %w", err)
		}
		prepared.PublicOCR.Provider = detection.Provider
		prepared.PublicOCR.Protocol = detection.Protocol
	}
	if allow && shouldDetectVerificationProvider(current.Verification, prepared.Verification) {
		probeConfiguration := prepared
		probeConfiguration.Verification.Provider = ""
		probeConfiguration.Verification.Protocol = ""
		detection, err := invoiceProvider.TestConnection(ctx, "verification", probeConfiguration)
		if err != nil {
			return config.InvoiceRecognition{}, fmt.Errorf("权威验真供应商自动探测失败: %w", err)
		}
		prepared.Verification.Provider = detection.Provider
		prepared.Verification.Protocol = detection.Protocol
	}
	if allow && shouldDetectMultimodalProtocol(current.Multimodal, prepared.Multimodal) {
		probeConfiguration := prepared
		probeConfiguration.Multimodal.Protocol = ""
		if err := probeConfiguration.Validate(); err != nil {
			return config.InvoiceRecognition{}, err
		}
		detection, err := invoiceProvider.TestConnection(ctx, "multimodal", probeConfiguration)
		if err != nil {
			return config.InvoiceRecognition{}, fmt.Errorf("多模态接口协议自动探测失败: %w", err)
		}
		prepared.Multimodal.Protocol = detection.Protocol
	}
	if err := prepared.Validate(); err != nil {
		return config.InvoiceRecognition{}, err
	}
	return prepared, nil
}

func shouldDetectPublicOCR(current config.InvoicePublicOCR, next config.InvoicePublicOCR) bool {
	if !next.Enabled {
		return false
	}
	detectionSupported := next.Provider == config.OCRProviderHTTPCompatible &&
		next.Protocol == config.OCRProtocolMultipartJSONV1
	return !detectionSupported || !current.Enabled || next.Provider != current.Provider ||
		next.Protocol != current.Protocol || next.Endpoint != current.Endpoint || next.APIKey != current.APIKey
}

func shouldDetectVerificationProvider(
	current config.InvoiceVerificationProvider,
	next config.InvoiceVerificationProvider,
) bool {
	if !next.Enabled {
		return false
	}
	detectionSupported := (next.Provider == config.VerificationProviderBaidu &&
		next.Protocol == config.VerificationProtocolBaiduVATV1) ||
		(next.Provider == config.VerificationProviderHTTPCompatible &&
			next.Protocol == config.VerificationProtocolHTTPJSONV1)
	return !detectionSupported || !current.Enabled || next.Provider != current.Provider ||
		next.Protocol != current.Protocol || next.Endpoint != current.Endpoint ||
		next.APIKey != current.APIKey || next.SecretKey != current.SecretKey
}

func shouldDetectMultimodalProtocol(
	current config.InvoiceMultimodalProvider,
	next config.InvoiceMultimodalProvider,
) bool {
	if !next.Enabled {
		return false
	}
	protocolSupported := next.Protocol == config.MultimodalProtocolOpenAICompatible ||
		next.Protocol == config.MultimodalProtocolAnthropic
	return !protocolSupported || !current.Enabled || next.Protocol != current.Protocol ||
		next.BaseURL != current.BaseURL || next.Model != current.Model || next.APIKey != current.APIKey
}

//@author: [SliverHorn](https://github.com/SliverHorn)
//@function: GetServerInfo
//@description: 获取服务器信息
//@return: server *utils.Server, err error

func (systemConfigService *SystemConfigService) GetServerInfo() (server *utils.Server, err error) {
	var s utils.Server
	s.CollectedAt = time.Now()
	s.Os = utils.InitOS()
	if s.Host, err = utils.InitHost(); err != nil {
		global.GVA_LOG.Error("func utils.InitHost() Failed", zap.String("err", err.Error()))
		return &s, err
	}
	if s.Load, err = utils.InitLoad(); err != nil {
		global.GVA_LOG.Error("func utils.InitLoad() Failed", zap.String("err", err.Error()))
		return &s, err
	}
	if s.Cpu, err = utils.InitCPU(); err != nil {
		global.GVA_LOG.Error("func utils.InitCPU() Failed", zap.String("err", err.Error()))
		return &s, err
	}
	if s.Ram, err = utils.InitRAM(); err != nil {
		global.GVA_LOG.Error("func utils.InitRAM() Failed", zap.String("err", err.Error()))
		return &s, err
	}
	if s.Disk, err = utils.InitDisk(); err != nil {
		global.GVA_LOG.Error("func utils.InitDisk() Failed", zap.String("err", err.Error()))
		return &s, err
	}

	return &s, nil
}

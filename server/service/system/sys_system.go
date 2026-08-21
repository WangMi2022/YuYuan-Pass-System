package system

import (
	"context"
	"fmt"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/config"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/system"
	invoiceProvider "github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/provider"
	"github.com/WangMi2022/mit-assets-admin/server/utils"
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
	global.GVA_CONFIG_LOCK.Lock()
	configuration := global.GVA_CONFIG
	global.GVA_CONFIG_LOCK.Unlock()
	conf, configuredSecrets = redactSystemConfigSecrets(configuration)
	return conf, configuredSecrets, nil
}

func (systemConfigService *SystemConfigService) GetSystemConfigSecret(path string) (string, error) {
	global.GVA_CONFIG_LOCK.Lock()
	configuration := global.GVA_CONFIG
	global.GVA_CONFIG_LOCK.Unlock()
	value, ok := revealSystemConfigSecret(configuration, path)
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
	global.GVA_CONFIG_LOCK.Lock()
	defer global.GVA_CONFIG_LOCK.Unlock()
	system.Config.AI, err = prepareAIConfig(system.Config.AI, global.GVA_CONFIG.AI, allowInvoiceRecognition)
	if err != nil {
		return InvoiceRecognitionDetections{}, err
	}
	// 发票智能配置由 AI 服务管理专用接口维护，不再参与运行配置整体保存。
	system.Config.AI.Invoice = global.GVA_CONFIG.AI.Invoice
	mergeSystemConfigSecrets(&system.Config, global.GVA_CONFIG)
	system.Config.Captcha.Normalize()
	if err = system.Config.Captcha.Validate(); err != nil {
		return InvoiceRecognitionDetections{}, err
	}
	system.Config.ContactVerification.Normalize()
	if err = system.Config.ContactVerification.Validate(system.Config.Email); err != nil {
		return InvoiceRecognitionDetections{}, err
	}
	cs := utils.StructToMap(system.Config)
	for k, v := range cs {
		global.GVA_VP.Set(k, v)
	}
	if err = global.GVA_VP.WriteConfig(); err != nil {
		return InvoiceRecognitionDetections{}, err
	}
	// The invoice worker reads this section for every job, so provider changes
	// take effect immediately without restarting database connections.
	global.GVA_CONFIG.AI = system.Config.AI
	// Contact verification reads these sections per request. Applying them
	// immediately keeps the capability endpoint and sender in sync after save.
	global.GVA_CONFIG.Email = system.Config.Email
	global.GVA_CONFIG.ContactVerification = system.Config.ContactVerification
	// Login challenge generation and verification read this section per request.
	global.GVA_CONFIG.Captcha = system.Config.Captcha
	invoiceProvider.SetRuntimeInvoiceRecognition(system.Config.AI.Invoice)
	return invoiceRecognitionDetections(system.Config.AI.Invoice), nil
}

func (systemConfigService *SystemConfigService) GetInvoiceRecognitionConfig() config.InvoiceRecognition {
	global.GVA_CONFIG_LOCK.Lock()
	configuration := global.GVA_CONFIG.AI.Invoice
	global.GVA_CONFIG_LOCK.Unlock()
	configuration.Normalize()
	return configuration.Redacted()
}

func (systemConfigService *SystemConfigService) SetInvoiceRecognitionConfig(ctx context.Context, incoming config.InvoiceRecognition) (InvoiceRecognitionDetections, error) {
	global.GVA_CONFIG_LOCK.Lock()
	defer global.GVA_CONFIG_LOCK.Unlock()
	prepared, err := prepareInvoiceRecognitionConfig(ctx, incoming, global.GVA_CONFIG.AI.Invoice, true)
	if err != nil {
		return InvoiceRecognitionDetections{}, err
	}
	global.GVA_VP.Set("ai.invoice", prepared)
	if err := global.GVA_VP.WriteConfig(); err != nil {
		return InvoiceRecognitionDetections{}, err
	}
	global.GVA_CONFIG.AI.Invoice = prepared
	invoiceProvider.SetRuntimeInvoiceRecognition(prepared)
	return invoiceRecognitionDetections(prepared), nil
}

func (systemConfigService *SystemConfigService) TestInvoiceRecognitionProvider(ctx context.Context, target string, incoming config.InvoiceRecognition) (invoiceProvider.ConnectionInfo, error) {
	global.GVA_CONFIG_LOCK.Lock()
	current := global.GVA_CONFIG.AI.Invoice
	global.GVA_CONFIG_LOCK.Unlock()
	prepared := incoming.MergeSecrets(current, true)
	prepared.Normalize()
	return invoiceProvider.TestConnection(ctx, target, prepared)
}

func prepareAIConfig(incoming, current config.AI, allow bool) (config.AI, error) {
	prepared := incoming.MergeSecrets(current, allow)
	prepared.Invoice = current.Invoice
	prepared.Normalize()
	if err := prepared.Validate(); err != nil {
		return config.AI{}, err
	}
	return prepared, nil
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

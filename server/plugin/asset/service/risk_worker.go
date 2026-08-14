package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var riskWorkerOnce sync.Once

func (s *riskService) StartWorker() {
	riskWorkerOnce.Do(func() {
		if global.GVA_Timer == nil {
			global.GVA_LOG.Warn("定时任务服务不可用，跳过资产风险扫描任务")
			return
		}
		run := func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			resumeRunID, err := s.recoverableScanRun(ctx)
			if err != nil {
				global.GVA_LOG.Error("读取可恢复资产风险扫描失败", zap.Error(err))
				return
			}
			if _, err := s.StartScan(model.RiskScanTriggerScheduled, 0, "系统定时任务", resumeRunID); err != nil && !errors.Is(err, errRiskScanActive) {
				global.GVA_LOG.Error("启动资产风险定时扫描失败", zap.Error(err))
			}
		}
		if _, err := global.GVA_Timer.AddTaskByFunc(
			"AssetRiskDailyScan", "15 2 * * *", run, "每日资产风险扫描",
		); err != nil {
			global.GVA_LOG.Error("注册资产风险定时任务失败", zap.Error(err))
			return
		}
		go run()
	})
}

func (s *riskService) recoverableScanRun(ctx context.Context) (uint, error) {
	var running model.AssetRiskScanRun
	err := global.GVA_DB.WithContext(ctx).Where("status = ?", model.RiskScanStatusRunning).Order("started_at DESC").First(&running).Error
	if err == nil {
		if running.HeartbeatAt.After(time.Now().Add(-riskScanHeartbeatTimeout)) {
			return 0, errors.New("已有资产风险扫描正在运行")
		}
		return running.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	var failed model.AssetRiskScanRun
	err = global.GVA_DB.WithContext(ctx).Where("status = ?", model.RiskScanStatusFailed).Order("started_at DESC").First(&failed).Error
	if err == nil {
		if failed.ResumeCount >= 3 {
			return 0, nil
		}
		return failed.ID, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	return 0, err
}

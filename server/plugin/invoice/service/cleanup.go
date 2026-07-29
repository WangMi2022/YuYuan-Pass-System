package service

import (
	"context"
	"errors"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormLogger "gorm.io/gorm/logger"
)

func runInvoiceFileCleanupWorker() {
	if global.GVA_DB == nil {
		return
	}
	recoverStaleInvoiceFileCleanupJobs()
	processInvoiceFileCleanupBatch()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		recoverStaleInvoiceFileCleanupJobs()
		processInvoiceFileCleanupBatch()
	}
}

func recoverStaleInvoiceFileCleanupJobs() {
	staleBefore := time.Now().Add(-recognitionLeaseTimeout)
	if err := global.GVA_DB.Model(&model.InvoiceFileCleanupJob{}).
		Where("status = ? AND (locked_at IS NULL OR locked_at < ?)", model.FileCleanupJobProcessing, staleBefore).
		Updates(map[string]any{
			"status": model.FileCleanupJobPending, "locked_at": nil,
			"lock_token": "", "next_run_at": time.Now(),
		}).Error; err != nil && global.GVA_LOG != nil {
		global.GVA_LOG.Warn("恢复发票文件清理任务失败", zap.Error(err))
	}
}

func processInvoiceFileCleanupBatch() {
	for processed := 0; processed < 4; processed++ {
		job, err := claimInvoiceFileCleanupJob(0)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) && global.GVA_LOG != nil {
				global.GVA_LOG.Error("领取发票文件清理任务失败", zap.Error(err))
			}
			return
		}
		executeInvoiceFileCleanupJob(job)
	}
}

func processInvoiceFileCleanupJob(id uint) {
	job, err := claimInvoiceFileCleanupJob(id)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) && global.GVA_LOG != nil {
			global.GVA_LOG.Warn("领取指定发票文件清理任务失败", zap.Error(err), zap.Uint("cleanupJobID", id))
		}
		return
	}
	executeInvoiceFileCleanupJob(job)
}

func claimInvoiceFileCleanupJob(id uint) (model.InvoiceFileCleanupJob, error) {
	var claimed model.InvoiceFileCleanupJob
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var candidate model.InvoiceFileCleanupJob
		now := time.Now()
		query := tx.Session(&gorm.Session{Logger: gormLogger.Discard}).
			Where("status = ? AND (next_run_at IS NULL OR next_run_at <= ?)", model.FileCleanupJobPending, now).
			Order("next_run_at ASC, id ASC")
		if id != 0 {
			query = query.Where("id = ?", id)
		}
		if tx.Dialector.Name() == "postgres" {
			query = query.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})
		}
		if err := query.First(&candidate).Error; err != nil {
			return err
		}
		lockToken := uuid.NewString()
		result := tx.Model(&model.InvoiceFileCleanupJob{}).
			Where("id = ? AND status = ?", candidate.ID, model.FileCleanupJobPending).
			Updates(map[string]any{
				"status": model.FileCleanupJobProcessing, "attempts": gorm.Expr("attempts + 1"),
				"locked_at": &now, "lock_token": lockToken,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.First(&claimed, candidate.ID).Error
	})
	return claimed, err
}

func executeInvoiceFileCleanupJob(job model.InvoiceFileCleanupJob) {
	invoice := model.Invoice{
		FileKey: job.FileKey, StorageType: job.StorageType,
		StorageRoot: job.StorageRoot, StorageEndpoint: job.StorageEndpoint,
		StorageBucket: job.StorageBucket, StorageUseSSL: job.StorageUseSSL,
	}
	if err := deleteInvoiceObject(context.Background(), invoice); err != nil {
		finishInvoiceFileCleanupFailure(job, err)
		return
	}
	result := global.GVA_DB.Unscoped().
		Where("id = ? AND status = ? AND lock_token = ?", job.ID, model.FileCleanupJobProcessing, job.LockToken).
		Delete(&model.InvoiceFileCleanupJob{})
	if result.Error != nil && global.GVA_LOG != nil {
		global.GVA_LOG.Error("完成发票文件清理任务失败", zap.Error(result.Error), zap.Uint("cleanupJobID", job.ID))
	} else if result.Error == nil && result.RowsAffected == 0 && global.GVA_LOG != nil {
		global.GVA_LOG.Warn("发票文件已清理但任务租约已失效", zap.Uint("cleanupJobID", job.ID))
	}
}

func finishInvoiceFileCleanupFailure(job model.InvoiceFileCleanupJob, cleanupErr error) {
	message := truncateUTF8(cleanupErr.Error(), 1000)
	delay := 30 * time.Second
	for attempt := 1; attempt < job.Attempts && delay < time.Hour; attempt++ {
		delay *= 2
	}
	if delay > time.Hour {
		delay = time.Hour
	}
	nextRunAt := time.Now().Add(delay)
	result := global.GVA_DB.Model(&model.InvoiceFileCleanupJob{}).
		Where("id = ? AND status = ? AND lock_token = ?", job.ID, model.FileCleanupJobProcessing, job.LockToken).
		Updates(map[string]any{
			"status": model.FileCleanupJobPending, "locked_at": nil, "lock_token": "",
			"next_run_at": &nextRunAt, "last_error": message,
		})
	if result.Error != nil && global.GVA_LOG != nil {
		global.GVA_LOG.Error("记录发票文件清理失败状态失败", zap.Error(result.Error), zap.Uint("cleanupJobID", job.ID))
	}
}

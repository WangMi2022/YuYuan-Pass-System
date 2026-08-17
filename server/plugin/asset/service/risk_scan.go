package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	announcementService "github.com/WangMi2022/mit-assets-admin/server/plugin/announcement/service"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/asset/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const riskScanBatchSize = 200

const riskScanHeartbeatTimeout = 15 * time.Minute

const riskScanRetryDelay = 5 * time.Second

var riskScanActive atomic.Bool

var errRiskScanActive = errors.New("已有资产风险扫描正在运行")

type riskBatchResult struct {
	newEvents     int64
	updatedEvents int64
	notifications []model.AssetRiskEvent
}

func (s *riskService) StartScan(triggerType string, userID uint, userName string, resumeRunID uint) (model.AssetRiskScanRun, error) {
	if triggerType != model.RiskScanTriggerManual && triggerType != model.RiskScanTriggerScheduled {
		return model.AssetRiskScanRun{}, errors.New("风险扫描触发方式不正确")
	}
	if !riskScanActive.CompareAndSwap(false, true) {
		return model.AssetRiskScanRun{}, errRiskScanActive
	}
	run, err := s.prepareScanRun(triggerType, userID, userName, resumeRunID)
	if err != nil {
		riskScanActive.Store(false)
		return model.AssetRiskScanRun{}, err
	}
	go func(runID uint) {
		defer riskScanActive.Store(false)
		ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
		defer cancel()
		if err := s.runScan(ctx, runID); err != nil {
			if global.GVA_LOG != nil {
				global.GVA_LOG.Error("资产风险扫描失败", zap.Uint("runId", runID), zap.Error(err))
			}
			s.scheduleFailedScanRetry(runID)
		}
	}(run.ID)
	return run, nil
}

func (s *riskService) prepareScanRun(triggerType string, userID uint, userName string, resumeRunID uint) (model.AssetRiskScanRun, error) {
	now := time.Now()
	if err := s.expireStaleScanRuns(now); err != nil {
		return model.AssetRiskScanRun{}, err
	}
	var activeCount int64
	activeQuery := global.GVA_DB.Model(&model.AssetRiskScanRun{}).
		Where("status = ? AND heartbeat_at >= ?", model.RiskScanStatusRunning, now.Add(-riskScanHeartbeatTimeout))
	if resumeRunID != 0 {
		activeQuery = activeQuery.Where("id <> ?", resumeRunID)
	}
	if err := activeQuery.Count(&activeCount).Error; err != nil {
		return model.AssetRiskScanRun{}, err
	}
	if activeCount > 0 {
		return model.AssetRiskScanRun{}, errRiskScanActive
	}
	if resumeRunID == 0 {
		run := model.AssetRiskScanRun{
			TriggerType: triggerType, Status: model.RiskScanStatusRunning,
			StartedAt: now, HeartbeatAt: now, TriggeredBy: userID, TriggeredByName: strings.TrimSpace(userName),
		}
		if err := global.GVA_DB.Create(&run).Error; err != nil {
			if isRiskScanConflict(err) {
				return model.AssetRiskScanRun{}, errRiskScanActive
			}
			return model.AssetRiskScanRun{}, err
		}
		return run, nil
	}
	var run model.AssetRiskScanRun
	if err := global.GVA_DB.First(&run, resumeRunID).Error; err != nil {
		return model.AssetRiskScanRun{}, errors.New("待续扫任务不存在")
	}
	var latest model.AssetRiskScanRun
	if err := global.GVA_DB.Order("started_at DESC, id DESC").First(&latest).Error; err != nil {
		return model.AssetRiskScanRun{}, err
	}
	if latest.ID != run.ID {
		return model.AssetRiskScanRun{}, errors.New("只能续扫最新的未完成扫描任务")
	}
	if run.Status == model.RiskScanStatusSuccess {
		return model.AssetRiskScanRun{}, errors.New("已完成的扫描任务不能续扫")
	}
	if run.Status == model.RiskScanStatusRunning && run.HeartbeatAt.After(now.Add(-riskScanHeartbeatTimeout)) {
		return model.AssetRiskScanRun{}, errRiskScanActive
	}
	run.TriggerType = triggerType
	run.Status = model.RiskScanStatusRunning
	run.HeartbeatAt = now
	run.ErrorMessage = ""
	run.FinishedAt = nil
	run.ResumeCount++
	if userID != 0 {
		run.TriggeredBy = userID
		run.TriggeredByName = strings.TrimSpace(userName)
	}
	err := global.GVA_DB.Model(&run).Updates(map[string]any{
		"trigger_type": run.TriggerType, "status": run.Status, "heartbeat_at": run.HeartbeatAt,
		"error_message": "", "finished_at": nil, "triggered_by": run.TriggeredBy,
		"triggered_by_name": run.TriggeredByName, "resume_count": run.ResumeCount,
	}).Error
	if isRiskScanConflict(err) {
		return model.AssetRiskScanRun{}, errRiskScanActive
	}
	return run, err
}

func (s *riskService) expireStaleScanRuns(now time.Time) error {
	return global.GVA_DB.Model(&model.AssetRiskScanRun{}).
		Where("status = ? AND heartbeat_at < ?", model.RiskScanStatusRunning, now.Add(-riskScanHeartbeatTimeout)).
		Updates(map[string]any{
			"status":        model.RiskScanStatusFailed,
			"finished_at":   now,
			"heartbeat_at":  now,
			"error_message": "扫描心跳超时，等待续扫",
		}).Error
}

func (s *riskService) runScan(ctx context.Context, runID uint) (err error) {
	var run model.AssetRiskScanRun
	if err = global.GVA_DB.WithContext(ctx).First(&run, runID).Error; err != nil {
		return err
	}
	defer func() {
		if err != nil {
			s.markScanFailed(runID, err)
		}
	}()

	var rules []model.AssetRiskRule
	if err = global.GVA_DB.WithContext(ctx).Where("enabled = ?", true).Order("code ASC").Find(&rules).Error; err != nil {
		return err
	}
	serialGroups, similarByAsset, err := s.buildIdentityIndexes(ctx, rules)
	if err != nil {
		return err
	}

	cursor := run.CursorAssetID
	for {
		if err = ctx.Err(); err != nil {
			return err
		}
		var assets []model.Asset
		if err = global.GVA_DB.WithContext(ctx).Preload("Category").Where("id > ?", cursor).
			Order("id ASC").Limit(riskScanBatchSize).Find(&assets).Error; err != nil {
			return err
		}
		if len(assets) == 0 {
			break
		}
		assetIDs := make([]uint, 0, len(assets))
		for _, asset := range assets {
			assetIDs = append(assetIDs, asset.ID)
		}
		var records []model.AssetOperationRecord
		if err = global.GVA_DB.WithContext(ctx).Where("asset_id IN ?", assetIDs).
			Order("asset_id ASC, operated_at DESC, id DESC").Find(&records).Error; err != nil {
			return err
		}
		recordsByAsset := make(map[uint][]model.AssetOperationRecord, len(assets))
		for _, record := range records {
			recordsByAsset[record.AssetID] = append(recordsByAsset[record.AssetID], record)
		}
		evaluation := riskEvaluationContext{now: run.StartedAt, recordsByAsset: recordsByAsset, serialGroups: serialGroups, similarByAsset: similarByAsset}
		matches := make([]riskMatch, 0)
		for _, asset := range assets {
			for _, rule := range rules {
				matches = append(matches, evaluateRiskRule(rule, asset, evaluation)...)
			}
		}
		result, batchErr := s.persistRiskBatch(ctx, runID, assets, matches)
		for _, event := range result.notifications {
			announcementService.NotificationHub.Publish(announcementService.NotificationEvent{ID: event.ID, Title: event.Title, Kind: "asset-risk"})
		}
		if batchErr != nil {
			return batchErr
		}
		cursor = assets[len(assets)-1].ID
	}

	closed, err := s.closeStaleRisks(ctx, runID)
	if err != nil {
		return err
	}
	finishedAt := time.Now()
	return global.GVA_DB.WithContext(ctx).Model(&model.AssetRiskScanRun{}).Where("id = ?", runID).Updates(map[string]any{
		"status": model.RiskScanStatusSuccess, "finished_at": finishedAt, "heartbeat_at": finishedAt,
		"closed_events": gorm.Expr("closed_events + ?", closed), "error_message": "",
	}).Error
}

func (s *riskService) persistRiskBatch(ctx context.Context, runID uint, assets []model.Asset, matches []riskMatch) (riskBatchResult, error) {
	result := riskBatchResult{notifications: make([]model.AssetRiskEvent, 0)}
	matchesByAsset := make(map[uint][]riskMatch, len(assets))
	for _, match := range matches {
		matchesByAsset[match.asset.ID] = append(matchesByAsset[match.asset.ID], match)
	}

	for _, asset := range assets {
		assetResult, err := s.persistRiskAsset(ctx, runID, asset.ID, matchesByAsset[asset.ID])
		result.newEvents += assetResult.newEvents
		result.updatedEvents += assetResult.updatedEvents
		result.notifications = append(result.notifications, assetResult.notifications...)
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func (s *riskService) persistRiskAsset(ctx context.Context, runID, assetID uint, matches []riskMatch) (riskBatchResult, error) {
	result := riskBatchResult{notifications: make([]model.AssetRiskEvent, 0)}
	now := time.Now()
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, match := range matches {
			created, updated, notify, err := upsertRiskMatch(tx, runID, now, match)
			if err != nil {
				return err
			}
			if created {
				result.newEvents++
			}
			if updated {
				result.updatedEvents++
			}
			if notify != nil {
				result.notifications = append(result.notifications, *notify)
			}
		}
		return tx.Model(&model.AssetRiskScanRun{}).Where("id = ?", runID).Updates(map[string]any{
			"cursor_asset_id": assetID, "scanned_assets": gorm.Expr("scanned_assets + 1"),
			"new_events":     gorm.Expr("new_events + ?", result.newEvents),
			"updated_events": gorm.Expr("updated_events + ?", result.updatedEvents), "heartbeat_at": now,
		}).Error
	})
	return result, err
}

func upsertRiskMatch(tx *gorm.DB, runID uint, now time.Time, match riskMatch) (created, updated bool, notify *model.AssetRiskEvent, err error) {
	fingerprint := riskFingerprint(match.asset.ID, match.rule.Code, match.rule.Version, match.evidenceKey)
	var event model.AssetRiskEvent
	err = tx.Where("fingerprint = ?", fingerprint).First(&event).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		event = model.AssetRiskEvent{
			Fingerprint: fingerprint, AssetID: match.asset.ID, RuleCode: match.rule.Code, RuleVersion: match.rule.Version,
			Category: match.rule.Category, Severity: match.rule.Severity, Status: model.RiskStatusOpen,
			Title: match.title, Description: match.description, Evidence: match.evidence,
			Recommendation: match.recommendation, FirstDetectedAt: now, LastDetectedAt: now, LastScanRunID: runID,
		}
		if err = tx.Create(&event).Error; err != nil {
			return false, false, nil, err
		}
		if err = createRiskLog(tx, event.ID, "detected", "", model.RiskStatusOpen, 0, "系统扫描", "首次命中风险规则"); err != nil {
			return false, false, nil, err
		}
		if isHighRisk(event.Severity) {
			notify = &event
		}
		return true, false, notify, nil
	}
	if err != nil {
		return false, false, nil, err
	}
	updates := map[string]any{
		"category": match.rule.Category, "severity": match.rule.Severity, "title": match.title,
		"description": match.description, "evidence": match.evidence, "recommendation": match.recommendation,
		"last_detected_at": now, "last_scan_run_id": runID,
	}
	if event.Status == model.RiskStatusResolved {
		updates["status"] = model.RiskStatusOpen
		updates["handled_by"] = 0
		updates["handled_by_name"] = ""
		updates["handled_at"] = nil
		updates["resolution_note"] = ""
		if err = createRiskLog(tx, event.ID, "auto_reopen", event.Status, model.RiskStatusOpen, 0, "系统扫描", "风险再次命中，自动重新打开"); err != nil {
			return false, false, nil, err
		}
		event.Status = model.RiskStatusOpen
		if isHighRisk(match.rule.Severity) {
			notify = &event
			notify.Status = model.RiskStatusOpen
			notify.Title = match.title
			notify.Severity = match.rule.Severity
		}
	}
	err = tx.Model(&event).Updates(updates).Error
	return false, true, notify, err
}

func (s *riskService) closeStaleRisks(ctx context.Context, runID uint) (int64, error) {
	var total int64
	for {
		var events []model.AssetRiskEvent
		if err := global.GVA_DB.WithContext(ctx).Where("status IN ? AND last_scan_run_id <> ?", []string{model.RiskStatusOpen, model.RiskStatusAcknowledged}, runID).
			Order("id ASC").Limit(riskScanBatchSize).Find(&events).Error; err != nil {
			return total, err
		}
		if len(events) == 0 {
			return total, nil
		}
		for _, candidate := range events {
			now := time.Now()
			resolved := false
			if err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				var event model.AssetRiskEvent
				err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status IN ?", candidate.ID, []string{model.RiskStatusOpen, model.RiskStatusAcknowledged}).First(&event).Error
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				if err != nil {
					return err
				}
				result := tx.Model(&event).Where("status IN ?", []string{model.RiskStatusOpen, model.RiskStatusAcknowledged}).Updates(map[string]any{
					"status": model.RiskStatusResolved, "handled_at": now, "handled_by": 0,
					"handled_by_name": "系统扫描", "resolution_note": "本次扫描已不再命中该风险",
				})
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected != 1 {
					return nil
				}
				if err := createRiskLog(tx, event.ID, "auto_resolve", event.Status, model.RiskStatusResolved, 0, "系统扫描", "本次扫描已不再命中该风险"); err != nil {
					return err
				}
				resolved = true
				return nil
			}); err != nil {
				return total, err
			}
			if resolved {
				total++
			}
		}
	}
}

func isRiskScanConflict(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(message, "idx_asset_risk_scan_single_running") || strings.Contains(message, "unique constraint failed: asset_risk_scan_runs.status")
}

func (s *riskService) scheduleFailedScanRetry(runID uint) {
	time.AfterFunc(riskScanRetryDelay, func() {
		var run model.AssetRiskScanRun
		if err := global.GVA_DB.First(&run, runID).Error; err != nil {
			if global.GVA_LOG != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				global.GVA_LOG.Error("读取失败资产风险扫描任务失败", zap.Uint("runId", runID), zap.Error(err))
			}
			return
		}
		if run.Status != model.RiskScanStatusFailed || run.ResumeCount >= 3 {
			return
		}
		if _, err := s.StartScan(model.RiskScanTriggerScheduled, 0, "系统自动续扫", run.ID); err != nil && !errors.Is(err, errRiskScanActive) && global.GVA_LOG != nil {
			global.GVA_LOG.Error("启动资产风险自动续扫失败", zap.Uint("runId", run.ID), zap.Error(err))
		}
	})
}

func (s *riskService) buildIdentityIndexes(ctx context.Context, rules []model.AssetRiskRule) (map[string][]similarAsset, map[uint][]similarAsset, error) {
	var identities []struct {
		ID           uint
		AssetCode    string
		SerialNumber string
	}
	if err := global.GVA_DB.WithContext(ctx).Model(&model.Asset{}).Select("id", "asset_code", "serial_number").Order("id ASC").Find(&identities).Error; err != nil {
		return nil, nil, err
	}
	serialGroups := make(map[string][]similarAsset)
	items := make([]similarAsset, 0, len(identities))
	for _, identity := range identities {
		if serial := normalizeIdentity(identity.SerialNumber); serial != "" {
			serialGroups[serial] = append(serialGroups[serial], similarAsset{ID: identity.ID, Code: identity.AssetCode})
		}
		items = append(items, similarAsset{ID: identity.ID, Code: identity.AssetCode})
	}
	maxDistance, minLength := 1, 5
	for _, rule := range rules {
		if rule.Code == riskRuleSimilarCode {
			maxDistance = paramInt(rule.Parameters, "maxDistance", maxDistance)
			minLength = paramInt(rule.Parameters, "minLength", minLength)
			break
		}
	}
	return serialGroups, buildSimilarAssetIndex(items, maxDistance, minLength), nil
}

func buildSimilarAssetIndex(items []similarAsset, maxDistance, minLength int) map[uint][]similarAsset {
	result := make(map[uint][]similarAsset)
	var root *bkNode
	for _, item := range items {
		normalized := normalizeIdentity(item.Code)
		if len([]rune(normalized)) < minLength {
			continue
		}
		if root == nil {
			root = &bkNode{value: normalized, items: []similarAsset{item}, children: make(map[int]*bkNode)}
			continue
		}
		root.insert(normalized, item)
	}
	seen := make(map[[2]uint]struct{})
	for _, item := range items {
		normalized := normalizeIdentity(item.Code)
		if root == nil || len([]rune(normalized)) < minLength {
			continue
		}
		for _, candidate := range root.search(normalized, maxDistance) {
			if candidate.ID == item.ID {
				continue
			}
			pair := [2]uint{item.ID, candidate.ID}
			if pair[0] > pair[1] {
				pair[0], pair[1] = pair[1], pair[0]
			}
			if _, exists := seen[pair]; exists {
				continue
			}
			seen[pair] = struct{}{}
			result[item.ID] = append(result[item.ID], candidate)
			result[candidate.ID] = append(result[candidate.ID], item)
		}
	}
	for assetID := range result {
		sort.Slice(result[assetID], func(i, j int) bool { return result[assetID][i].ID < result[assetID][j].ID })
	}
	return result
}

type bkNode struct {
	value    string
	items    []similarAsset
	children map[int]*bkNode
}

func (n *bkNode) insert(value string, item similarAsset) {
	distance := levenshteinDistance(n.value, value)
	if distance == 0 {
		n.items = append(n.items, item)
		return
	}
	child := n.children[distance]
	if child == nil {
		n.children[distance] = &bkNode{value: value, items: []similarAsset{item}, children: make(map[int]*bkNode)}
		return
	}
	child.insert(value, item)
}

func (n *bkNode) search(value string, maxDistance int) []similarAsset {
	if n == nil {
		return nil
	}
	distance := levenshteinDistance(n.value, value)
	result := make([]similarAsset, 0)
	if distance <= maxDistance {
		result = append(result, n.items...)
	}
	for edge, child := range n.children {
		if edge >= distance-maxDistance && edge <= distance+maxDistance {
			result = append(result, child.search(value, maxDistance)...)
		}
	}
	return result
}

func levenshteinDistance(left, right string) int {
	a, b := []rune(left), []rune(right)
	if len(a) > len(b) {
		a, b = b, a
	}
	previous := make([]int, len(a)+1)
	for index := range previous {
		previous[index] = index
	}
	for row, rightRune := range b {
		current := make([]int, len(a)+1)
		current[0] = row + 1
		for column, leftRune := range a {
			cost := 0
			if leftRune != rightRune {
				cost = 1
			}
			current[column+1] = minInt(current[column]+1, previous[column+1]+1, previous[column]+cost)
		}
		previous = current
	}
	return previous[len(a)]
}

func minInt(values ...int) int {
	minimum := values[0]
	for _, value := range values[1:] {
		if value < minimum {
			minimum = value
		}
	}
	return minimum
}

func riskFingerprint(assetID uint, ruleCode string, ruleVersion int, evidenceKey string) string {
	value := fmt.Sprintf("%d|%s|%d|%s", assetID, ruleCode, ruleVersion, evidenceKey)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(value)))
}

func isHighRisk(severity string) bool {
	return severity == model.RiskSeverityHigh || severity == model.RiskSeverityCritical
}

func createRiskLog(tx *gorm.DB, eventID uint, action, fromState, toState string, actorID uint, actorName, note string) error {
	return tx.Create(&model.AssetRiskEventLog{
		EventID: eventID, Action: action, FromState: fromState, ToState: toState,
		ActorID: actorID, ActorName: strings.TrimSpace(actorName), Note: strings.TrimSpace(note),
	}).Error
}

func (s *riskService) markScanFailed(runID uint, scanErr error) {
	message := strings.TrimSpace(scanErr.Error())
	if len(message) > 2000 {
		message = message[:2000]
	}
	now := time.Now()
	_ = global.GVA_DB.Model(&model.AssetRiskScanRun{}).Where("id = ?", runID).Updates(map[string]any{
		"status": model.RiskScanStatusFailed, "finished_at": now, "heartbeat_at": now, "error_message": message,
	}).Error
}

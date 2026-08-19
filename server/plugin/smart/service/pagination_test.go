package service

import (
	"fmt"
	"testing"
	"time"

	commonRequest "github.com/WangMi2022/mit-assets-admin/server/model/common/request"
	smartModel "github.com/WangMi2022/mit-assets-admin/server/plugin/smart/model"
)

func TestPagedSmartCollectionsReturnAllReachableRecords(t *testing.T) {
	database := setupSmartTestDB(t)
	now := time.Now().Add(-time.Hour)

	for index := 0; index < 12; index++ {
		session := smartModel.CopilotSession{
			UserID: 1, AuthorityID: 888, Title: fmt.Sprintf("会话-%02d", index),
			LastMessageAt: now.Add(time.Duration(index) * time.Minute),
		}
		if err := database.Create(&session).Error; err != nil {
			t.Fatalf("create session %d: %v", index, err)
		}
		delivery := smartModel.SmartReportDelivery{
			UserID: 1, ReportID: uint(index + 1), Channel: "in_app", Status: "sent",
		}
		if err := database.Create(&delivery).Error; err != nil {
			t.Fatalf("create delivery %d: %v", index, err)
		}
		draft := smartModel.SmartDraft{
			UserID: 1, DraftType: smartModel.DraftTypeSchedule, Status: smartModel.DraftStatusDraft,
		}
		if err := database.Create(&draft).Error; err != nil {
			t.Fatalf("create draft %d: %v", index, err)
		}
	}
	if err := database.Create(&smartModel.CopilotSession{UserID: 2, AuthorityID: 888, Title: "其他用户", LastMessageAt: time.Now()}).Error; err != nil {
		t.Fatal(err)
	}

	firstPage := commonRequest.PageInfo{Page: 1, PageSize: 10}
	sessions, sessionTotal, err := Smart.SessionPage(1, 888, &firstPage)
	if err != nil || sessionTotal != 12 || len(sessions) != 10 {
		t.Fatalf("sessions page = %d/%d, total=%d, err=%v", len(sessions), firstPage.PageSize, sessionTotal, err)
	}
	secondPage := commonRequest.PageInfo{Page: 2, PageSize: 10}
	sessions, sessionTotal, err = Smart.SessionPage(1, 888, &secondPage)
	if err != nil || sessionTotal != 12 || len(sessions) != 2 {
		t.Fatalf("sessions second page = %d, total=%d, err=%v", len(sessions), sessionTotal, err)
	}

	deliveryPage := commonRequest.PageInfo{Page: 1, PageSize: 10}
	deliveries, deliveryTotal, err := Smart.DeliveryPage(1, &deliveryPage)
	if err != nil || deliveryTotal != 12 || len(deliveries) != 10 {
		t.Fatalf("deliveries page = %d, total=%d, err=%v", len(deliveries), deliveryTotal, err)
	}

	draftPage := commonRequest.PageInfo{Page: 2, PageSize: 10}
	drafts, draftTotal, err := Smart.DraftPage(1, smartModel.DraftTypeSchedule, &draftPage)
	if err != nil || draftTotal != 12 || len(drafts) != 2 {
		t.Fatalf("drafts second page = %d, total=%d, err=%v", len(drafts), draftTotal, err)
	}
}

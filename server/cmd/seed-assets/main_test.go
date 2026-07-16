package main

import "testing"

func TestBuildDraftSeedsCoversEveryOperation(t *testing.T) {
	statuses := []string{"pending_inbound", "idle", "in_use", "maintenance", "retired"}
	assets := make([]seededAsset, 0, 50)
	for index := 0; index < 50; index++ {
		status := statuses[index%len(statuses)]
		assets = append(assets, seededAsset{
			ID: int64(index + 1), Code: "TEST", Name: "测试资产", Quantity: 1,
			Status: status, Location: "当前位置", Custodian: "当前保管人",
		})
	}

	drafts, err := buildDraftSeeds(assets, 2)
	if err != nil {
		t.Fatalf("buildDraftSeeds returned error: %v", err)
	}
	if len(drafts) != 12 {
		t.Fatalf("got %d drafts, want 12", len(drafts))
	}

	counts := map[string]int{}
	usedAssets := map[int64]struct{}{}
	for _, draft := range drafts {
		counts[draft.Snapshot.OperationType]++
		if _, exists := usedAssets[draft.Asset.ID]; exists {
			t.Fatalf("asset %d was selected by multiple drafts", draft.Asset.ID)
		}
		usedAssets[draft.Asset.ID] = struct{}{}
		if !draftAssetEligible(draft.Snapshot.OperationType, draft.Asset.Status) {
			t.Fatalf("asset status %q is invalid for %q", draft.Asset.Status, draft.Snapshot.OperationType)
		}
	}
	for _, operationType := range []string{"inbound", "issue", "transfer", "return", "maintenance", "scrap"} {
		if counts[operationType] != 2 {
			t.Fatalf("operation %q got %d drafts, want 2", operationType, counts[operationType])
		}
	}
}

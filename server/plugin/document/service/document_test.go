package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	commonRequest "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/document/model"
	documentRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/document/model/request"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestListReturnsMetadataWithoutDocumentContent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:document-list-metadata?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err = db.AutoMigrate(&model.Document{}); err != nil {
		t.Fatalf("migrate documents: %v", err)
	}
	stored := model.Document{
		Title: "性能说明", OriginalName: "performance.md", FileExt: "md", FileSize: 2048,
		MimeType: "text/markdown", StorageType: "minio", FileKey: "documents/performance.md",
		FileURL: "https://storage.example/performance.md", Content: "confidential document content",
		Editable: true, Remarks: "internal note",
	}
	if err = db.Create(&stored).Error; err != nil {
		t.Fatalf("create document: %v", err)
	}

	previousDB := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = previousDB })

	list, total, err := Document.List(documentRequest.DocumentSearch{PageInfo: commonRequest.PageInfo{Page: 1, PageSize: 10}})
	if err != nil {
		t.Fatalf("list documents: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("list result = (%d, %d items), want (1, 1 item)", total, len(list))
	}
	if list[0].ID != stored.ID || list[0].Title != stored.Title || list[0].OriginalName != stored.OriginalName {
		t.Fatalf("metadata mismatch: %+v", list[0])
	}
	payload, err := json.Marshal(list[0])
	if err != nil {
		t.Fatalf("marshal list item: %v", err)
	}
	serialized := string(payload)
	for _, forbidden := range []string{"content", "remarks", "fileKey", "fileUrl", "confidential document content"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("list payload unexpectedly contains %q: %s", forbidden, serialized)
		}
	}
}

func TestUpdateContentRejectsOversizedDocument(t *testing.T) {
	_, err := Document.UpdateContent(documentRequest.UpdateContent{
		ID:      1,
		Content: strings.Repeat("a", MaxEditableBytes+1),
	})
	if err == nil || !strings.Contains(err.Error(), "2MB") {
		t.Fatalf("UpdateContent() error = %v, want 2MB limit error", err)
	}
}

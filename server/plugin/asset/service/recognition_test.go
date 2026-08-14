package service

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/ai"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"
	assetRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model/request"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type assetRecognitionGatewayStub struct {
	result ai.VisionResult
	err    error
}

func (stub assetRecognitionGatewayStub) Complete(context.Context, ai.CompletionRequest) (ai.CompletionResult, error) {
	return ai.CompletionResult{}, stub.err
}

func (stub assetRecognitionGatewayStub) Vision(_ context.Context, request ai.VisionRequest) (ai.VisionResult, error) {
	if request.Module != "asset" || request.PromptKey != assetRecognitionPromptKey || request.ObjectType != "asset_recognition_job" {
		return ai.VisionResult{}, &ai.Error{Type: ai.ErrorTypeValidation, Message: "unexpected asset recognition request"}
	}
	return stub.result, stub.err
}

func (stub assetRecognitionGatewayStub) Stream(context.Context, ai.CompletionRequest) (ai.StreamResult, error) {
	return ai.StreamResult{Response: &http.Response{}}, stub.err
}

func setupAssetRecognitionTestDB(t *testing.T) {
	t.Helper()
	previousDB := global.GVA_DB
	previousOssType := global.GVA_CONFIG.System.OssType
	previousStorePath := global.GVA_CONFIG.Local.StorePath
	previousLocalPath := global.GVA_CONFIG.Local.Path
	previousGateway := assetRecognitionGateway
	database, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open asset recognition test database: %v", err)
	}
	if err = database.AutoMigrate(
		&model.Category{}, &model.Asset{}, &model.AssetRecognitionJob{}, &ai.ModelInvocation{},
	); err != nil {
		t.Fatalf("migrate asset recognition tables: %v", err)
	}
	global.GVA_DB = database
	global.GVA_CONFIG.System.OssType = "local"
	global.GVA_CONFIG.Local.StorePath = t.TempDir()
	global.GVA_CONFIG.Local.Path = "/uploads/file"
	t.Cleanup(func() {
		assetRecognitionGateway = previousGateway
		global.GVA_DB = previousDB
		global.GVA_CONFIG.System.OssType = previousOssType
		global.GVA_CONFIG.Local.StorePath = previousStorePath
		global.GVA_CONFIG.Local.Path = previousLocalPath
	})
}

func createRecognitionPhotoHeader(t *testing.T) *multipart.FileHeader {
	t.Helper()
	var imageData bytes.Buffer
	assetImage := image.NewRGBA(image.Rect(0, 0, 2, 2))
	assetImage.Set(0, 0, color.RGBA{R: 24, G: 92, B: 180, A: 255})
	if err := png.Encode(&imageData, assetImage); err != nil {
		t.Fatalf("encode recognition image: %v", err)
	}
	var formData bytes.Buffer
	writer := multipart.NewWriter(&formData)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="files"; filename="nameplate.png"`)
	header.Set("Content-Type", "image/png")
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("create recognition multipart part: %v", err)
	}
	if _, err = part.Write(imageData.Bytes()); err != nil {
		t.Fatalf("write recognition image: %v", err)
	}
	if err = writer.Close(); err != nil {
		t.Fatalf("close recognition multipart writer: %v", err)
	}
	reader := multipart.NewReader(bytes.NewReader(formData.Bytes()), writer.Boundary())
	form, err := reader.ReadForm(1 << 20)
	if err != nil {
		t.Fatalf("read recognition multipart form: %v", err)
	}
	t.Cleanup(func() { _ = form.RemoveAll() })
	return form.File["files"][0]
}

func createAssetRecognitionCategory(t *testing.T) model.Category {
	t.Helper()
	category := model.Category{Name: "电脑整机", Code: "IT-COMPUTER", Enabled: true}
	if err := global.GVA_DB.Create(&category).Error; err != nil {
		t.Fatalf("create asset category: %v", err)
	}
	return category
}

func TestAssetRecognitionBuildsReviewDraftAndCreatesAssetOnce(t *testing.T) {
	setupAssetRecognitionTestDB(t)
	category := createAssetRecognitionCategory(t)
	assetRecognitionGateway = assetRecognitionGatewayStub{result: ai.VisionResult{
		Provider: "openai-compatible", Model: "vision-test",
		Content: `{"name":"研发笔记本","brand":"Example","model":"Pro 14","serialNumber":"SN-UNIQUE-001","specifications":"32GB RAM","productionDate":"2026-06-01","recommendedCategoryCode":"IT-COMPUTER","recommendedUnit":"台","recommendedWarrantyMonths":24,"rawText":"SN-UNIQUE-001","fieldConfidences":{"name":0.96,"brand":0.92,"model":0.91,"serialNumber":0.99,"specifications":0.68,"productionDate":0.87,"recommendedCategoryCode":0.93,"recommendedUnit":0.9,"recommendedWarrantyMonths":0.75}}`,
	}}
	job, err := (assetRecognitionService{}).Create([]*multipart.FileHeader{createRecognitionPhotoHeader(t)}, 10, 100)
	if err != nil {
		t.Fatalf("create asset recognition job: %v", err)
	}
	claimed, err := claimAssetRecognitionJob()
	if err != nil {
		t.Fatalf("claim asset recognition job: %v", err)
	}
	processAssetRecognitionJob(claimed)
	job, err = (assetRecognitionService{}).Get(job.ID, 10, 100)
	if err != nil {
		t.Fatalf("load reviewing recognition job: %v", err)
	}
	if job.Status != model.AssetRecognitionReviewing || job.Draft.Name != "研发笔记本" || job.Draft.CategoryID != category.ID {
		t.Fatalf("unexpected recognition draft: %#v", job)
	}
	if len(job.Warnings) == 0 || job.FieldConfidences["serialNumber"] != 0.99 {
		t.Fatalf("confidence warnings were not persisted: %#v", job)
	}
	job.Draft.AssetCode = "AI-ASSET-001"
	job.Draft.UnitPrice = 12000
	job.Draft.CurrentValue = 12000
	job, err = (assetRecognitionService{}).SaveDraft(assetRequest.AssetRecognitionDraftUpdate{ID: job.ID, Draft: job.Draft}, 10, 100)
	if err != nil {
		t.Fatalf("save asset recognition draft: %v", err)
	}
	created, err := (assetRecognitionService{}).Confirm(job.ID, 10, 100)
	if err != nil {
		t.Fatalf("confirm asset recognition job: %v", err)
	}
	if created.AssetCode != "AI-ASSET-001" || created.Specifications != "32GB RAM" || created.ProductionDate == nil {
		t.Fatalf("unexpected created asset: %#v", created)
	}
	if _, err = (assetRecognitionService{}).Confirm(job.ID, 10, 100); err == nil {
		t.Fatal("second confirmation should not create another asset")
	}
	var assetCount int64
	if err = global.GVA_DB.Model(&model.Asset{}).Count(&assetCount).Error; err != nil || assetCount != 1 {
		t.Fatalf("expected one formal asset, count=%d err=%v", assetCount, err)
	}
	job, err = (assetRecognitionService{}).Get(job.ID, 10, 100)
	if err != nil || job.Status != model.AssetRecognitionCompleted || job.ConfirmedAssetID == nil || *job.ConfirmedAssetID != created.ID {
		t.Fatalf("recognition job was not immutably linked: %#v err=%v", job, err)
	}
}

func TestParseAssetVisionOutputNormalizesControlledAliases(t *testing.T) {
	output, err := parseAssetVisionOutput(`{
		"productName":"  研发终端  ",
		"manufacturer":" Codex Lab ",
		"warrantyMonths":36,
		"fieldConfidences":{"name":0.1,"productName":0.96,"brand":0.2,"manufacturer":0.92,"recommendedWarrantyMonths":0.3,"warrantyMonths":0.81}
	}`)
	if err != nil {
		t.Fatalf("parse controlled aliases: %v", err)
	}
	if output.Name != "研发终端" || output.Brand != "Codex Lab" || output.RecommendedWarrantyMonths != 36 {
		t.Fatalf("aliases were not normalized: %#v", output)
	}
	if output.ProductName != "" || output.Manufacturer != "" || output.WarrantyMonths != 0 {
		t.Fatalf("aliases leaked after normalization: %#v", output)
	}
	if output.FieldConfidences["name"] != 0.96 || output.FieldConfidences["brand"] != 0.92 || output.FieldConfidences["recommendedWarrantyMonths"] != 0.81 {
		t.Fatalf("alias confidences were not normalized: %#v", output.FieldConfidences)
	}
	for _, alias := range []string{"productName", "manufacturer", "warrantyMonths"} {
		if _, exists := output.FieldConfidences[alias]; exists {
			t.Fatalf("alias confidence leaked after normalization: %s", alias)
		}
	}
}

func TestParseAssetVisionOutputPrefersCanonicalFields(t *testing.T) {
	output, err := parseAssetVisionOutput(`{
		"name":"标准名称",
		"productName":"别名名称",
		"brand":"标准品牌",
		"manufacturer":"别名厂商",
		"recommendedWarrantyMonths":24,
		"warrantyMonths":36,
		"fieldConfidences":{"name":0.91,"productName":0.99,"brand":0.88,"manufacturer":0.97,"recommendedWarrantyMonths":0.74,"warrantyMonths":0.95}
	}`)
	if err != nil {
		t.Fatalf("parse canonical and alias fields: %v", err)
	}
	if output.Name != "标准名称" || output.Brand != "标准品牌" || output.RecommendedWarrantyMonths != 24 {
		t.Fatalf("canonical fields should win: %#v", output)
	}
	if output.FieldConfidences["name"] != 0.91 || output.FieldConfidences["brand"] != 0.88 || output.FieldConfidences["recommendedWarrantyMonths"] != 0.74 {
		t.Fatalf("canonical confidences should win: %#v", output.FieldConfidences)
	}
}

func TestAssetRecognitionBlocksNormalizedDuplicateSerial(t *testing.T) {
	setupAssetRecognitionTestDB(t)
	category := createAssetRecognitionCategory(t)
	existing := model.Asset{
		AssetCode: "EXISTING-001", Name: "已有资产", CategoryID: category.ID,
		SerialNumber: "SN-2026-001", Quantity: 1, Unit: "台", UnitPrice: 100, CurrentValue: 100,
	}
	if err := Asset.Create(&existing); err != nil {
		t.Fatalf("create existing asset: %v", err)
	}
	job := model.AssetRecognitionJob{
		Status: model.AssetRecognitionReviewing, MaxAttempts: 3, CreatedBy: 20, AuthorityID: 200,
		StorageType: "local",
		Draft: model.AssetRecognitionDraft{
			AssetCode: "NEW-001", Name: "待确认资产", CategoryID: category.ID,
			SerialNumber: "sn 2026/001", Quantity: 1, Unit: "台", UnitPrice: 200, CurrentValue: 200,
		},
	}
	if err := global.GVA_DB.Create(&job).Error; err != nil {
		t.Fatalf("create reviewing recognition job: %v", err)
	}
	if _, err := (assetRecognitionService{}).Confirm(job.ID, 20, 200); err == nil {
		t.Fatal("normalized duplicate serial should block confirmation")
	}
	var assetCount int64
	if err := global.GVA_DB.Model(&model.Asset{}).Count(&assetCount).Error; err != nil || assetCount != 1 {
		t.Fatalf("duplicate confirmation created asset: count=%d err=%v", assetCount, err)
	}
}

func TestAssetRecognitionSchemaFailureBecomesRetryableFailedTask(t *testing.T) {
	setupAssetRecognitionTestDB(t)
	createAssetRecognitionCategory(t)
	assetRecognitionGateway = assetRecognitionGatewayStub{err: &ai.Error{Type: ai.ErrorTypeSchema, Message: "模型输出不是合法 JSON"}}
	job, err := (assetRecognitionService{}).Create([]*multipart.FileHeader{createRecognitionPhotoHeader(t)}, 30, 300)
	if err != nil {
		t.Fatalf("create schema failure job: %v", err)
	}
	claimed, err := claimAssetRecognitionJob()
	if err != nil {
		t.Fatalf("claim schema failure job: %v", err)
	}
	processAssetRecognitionJob(claimed)
	job, err = (assetRecognitionService{}).Get(job.ID, 30, 300)
	if err != nil {
		t.Fatalf("load failed recognition job: %v", err)
	}
	if job.Status != model.AssetRecognitionFailed || job.LastError == "" {
		t.Fatalf("schema failure did not enter failed state: %#v", job)
	}
	if err = (assetRecognitionService{}).Retry(job.ID, 30, 300); err != nil {
		t.Fatalf("failed schema job should be retryable: %v", err)
	}
}

func TestAssetRecognitionDeleteKeepsRetryableCleanupState(t *testing.T) {
	setupAssetRecognitionTestDB(t)
	root := global.GVA_CONFIG.Local.StorePath
	firstKey := "asset-delete-first.png"
	if err := os.WriteFile(filepath.Join(root, firstKey), []byte("first"), 0o600); err != nil {
		t.Fatal(err)
	}
	job := model.AssetRecognitionJob{
		Status: model.AssetRecognitionFailed, MaxAttempts: 3, CreatedBy: 40, AuthorityID: 400,
		StorageType: "local", StorageRoot: root,
		InputPhotos: []model.Photo{{Key: firstKey}, {Key: "../invalid.png"}},
	}
	if err := global.GVA_DB.Create(&job).Error; err != nil {
		t.Fatal(err)
	}
	if err := (assetRecognitionService{}).Delete(job.ID, 40, 400); err == nil {
		t.Fatal("partial image cleanup should report an error")
	}
	if _, err := os.Stat(filepath.Join(root, firstKey)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("first image was not removed: %v", err)
	}
	job, err := (assetRecognitionService{}).Get(job.ID, 40, 400)
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != model.AssetRecognitionDeleting || job.LockToken != "" || job.LastError == "" {
		t.Fatalf("cleanup failure did not remain retryable: %#v", job)
	}
	if err = (assetRecognitionService{}).Retry(job.ID, 40, 400); err == nil {
		t.Fatal("cleanup failure must not be requeued for recognition")
	}
	job.InputPhotos[1].Key = "already-removed.png"
	if err = global.GVA_DB.Save(&job).Error; err != nil {
		t.Fatal(err)
	}
	if err = (assetRecognitionService{}).Delete(job.ID, 40, 400); err != nil {
		t.Fatalf("retry cleanup: %v", err)
	}
	if _, err = (assetRecognitionService{}).Get(job.ID, 40, 400); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cleanup retry did not remove task: %v", err)
	}
}

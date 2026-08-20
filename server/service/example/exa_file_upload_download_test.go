package example

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	exampleModel "github.com/WangMi2022/mit-assets-admin/server/model/example"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func configureMediaPreviewTest(t *testing.T) {
	t.Helper()
	previous := global.GVA_CONFIG
	t.Cleanup(func() { global.GVA_CONFIG = previous })
	global.GVA_CONFIG.JWT.SigningKey = "media-preview-test-signing-key"
}

func TestAttachMediaPreviewURLsUsesSameOriginProxyWithoutReplacingCanonicalURLs(t *testing.T) {
	configureMediaPreviewTest(t)
	files := []exampleModel.ExaFileUploadAndDownload{
		{Url: "http://minio.local/gva-assets/assets/2026-08-17/photo.jpg", Key: "assets/2026-08-17/photo.jpg"},
		{Url: "https://cdn.example.com/imported.jpg", Key: "external-import-id"},
	}
	var signedKeys []string
	signer := func(_ context.Context, key string, expires time.Duration) (string, error) {
		if expires != mediaPreviewURLTTL {
			t.Fatalf("unexpected preview URL TTL: %s", expires)
		}
		signedKeys = append(signedKeys, key)
		return createMediaPreviewURL(context.Background(), key, expires)
	}

	if err := attachMediaPreviewURLs(context.Background(), files, "minio", "http://minio.local/gva-assets/", signer); err != nil {
		t.Fatalf("attach media preview URLs: %v", err)
	}
	if len(signedKeys) != 1 || signedKeys[0] != files[0].Key {
		t.Fatalf("unexpected signed keys: %#v", signedKeys)
	}
	if strings.Contains(files[0].PreviewURL, "minio.local") || !strings.HasPrefix(files[0].PreviewURL, "/api/fileUploadAndDownload/preview?") {
		t.Fatalf("managed object exposed its storage origin: %q", files[0].PreviewURL)
	}
	parsed, err := url.Parse(files[0].PreviewURL)
	if err != nil || parsed.Query().Get("key") != files[0].Key || !validateMediaPreviewToken(parsed.Query().Get("token"), files[0].Key) {
		t.Fatalf("managed object received an invalid application preview URL: %q, err=%v", files[0].PreviewURL, err)
	}
	if files[0].Url != "http://minio.local/gva-assets/assets/2026-08-17/photo.jpg" {
		t.Fatalf("canonical URL was replaced: %q", files[0].Url)
	}
	if files[1].PreviewURL != files[1].Url {
		t.Fatalf("external URL should remain unchanged: %q", files[1].PreviewURL)
	}
}

func TestAttachMediaPreviewURLsLeavesNonMinioStorageUntouched(t *testing.T) {
	files := []exampleModel.ExaFileUploadAndDownload{{Url: "/uploads/file/photo.jpg", Key: "photo.jpg"}}
	signer := func(context.Context, string, time.Duration) (string, error) {
		t.Fatal("non-MinIO storage must not call the signer")
		return "", nil
	}

	if err := attachMediaPreviewURLs(context.Background(), files, "local", "", signer); err != nil {
		t.Fatalf("attach local preview URL: %v", err)
	}
	if files[0].PreviewURL != files[0].Url {
		t.Fatalf("local preview URL should use the canonical URL: %q", files[0].PreviewURL)
	}
}

func TestAttachMediaPreviewURLsContinuesAfterOneObjectCannotBeSigned(t *testing.T) {
	configureMediaPreviewTest(t)
	files := []exampleModel.ExaFileUploadAndDownload{
		{Url: "http://minio.local/gva-assets/assets/broken.jpg", Key: "assets/broken.jpg"},
		{Url: "http://minio.local/gva-assets/assets/working.jpg", Key: "assets/working.jpg"},
	}
	signer := func(_ context.Context, key string, _ time.Duration) (string, error) {
		if key == "assets/broken.jpg" {
			return "", errors.New("test signing failure")
		}
		return createMediaPreviewURL(context.Background(), key, mediaPreviewURLTTL)
	}

	err := attachMediaPreviewURLs(context.Background(), files, "minio", "http://minio.local/gva-assets", signer)
	if err == nil {
		t.Fatal("expected the signing failure to be reported")
	}
	if files[0].PreviewURL != files[0].Url {
		t.Fatalf("failed object should fall back to its canonical URL: %q", files[0].PreviewURL)
	}
	if files[1].PreviewURL == files[1].Url {
		t.Fatalf("later objects must still be signed: %q", files[1].PreviewURL)
	}
}

func TestResolveMediaPreviewURLSignsPrivateAvatar(t *testing.T) {
	configureMediaPreviewTest(t)
	signer := func(_ context.Context, key string, expires time.Duration) (string, error) {
		if key != "avatars/user-7.png" {
			t.Fatalf("unexpected avatar object key: %q", key)
		}
		if expires != mediaPreviewURLTTL {
			t.Fatalf("unexpected preview URL TTL: %s", expires)
		}
		return createMediaPreviewURL(context.Background(), key, expires)
	}

	previewURL, err := resolveMediaPreviewURL(
		context.Background(),
		"http://minio.local/gva-assets/avatars/user-7.png",
		"minio",
		"http://minio.local/gva-assets/",
		signer,
	)
	if err != nil {
		t.Fatalf("resolve avatar preview URL: %v", err)
	}
	parsed, parseErr := url.Parse(previewURL)
	if parseErr != nil || parsed.Path != "/api/fileUploadAndDownload/preview" ||
		parsed.Query().Get("key") != "avatars/user-7.png" || !validateMediaPreviewToken(parsed.Query().Get("token"), "avatars/user-7.png") {
		t.Fatalf("private avatar did not receive a valid same-origin preview URL: %q, err=%v", previewURL, parseErr)
	}
}

func TestResolveMediaPreviewURLLeavesExternalAvatarUntouched(t *testing.T) {
	signer := func(context.Context, string, time.Duration) (string, error) {
		t.Fatal("external avatar must not call the MinIO signer")
		return "", nil
	}

	previewURL, err := resolveMediaPreviewURL(
		context.Background(),
		"https://cdn.example.com/avatar.png",
		"minio",
		"http://minio.local/gva-assets",
		signer,
	)
	if err != nil {
		t.Fatalf("resolve external avatar URL: %v", err)
	}
	if previewURL != "https://cdn.example.com/avatar.png" {
		t.Fatalf("external avatar URL changed: %q", previewURL)
	}
}

func TestResolveMediaPreviewObjectRequiresValidTokenAndManagedRecord(t *testing.T) {
	configureMediaPreviewTest(t)
	previousDB := global.GVA_DB
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open media preview database: %v", err)
	}
	if err = database.AutoMigrate(&exampleModel.ExaFileUploadAndDownload{}); err != nil {
		t.Fatalf("migrate media preview table: %v", err)
	}
	global.GVA_DB = database
	t.Cleanup(func() {
		global.GVA_DB = previousDB
		if sqlDB, dbErr := database.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
	})
	global.GVA_CONFIG.System.OssType = "minio"
	global.GVA_CONFIG.Minio.BucketUrl = "http://minio.internal/gva-assets"

	const key = "assets/2026-08-20/private-avatar.png"
	file := exampleModel.ExaFileUploadAndDownload{Name: "private-avatar.png", Key: key, Url: "http://minio.internal/gva-assets/" + key}
	if err = database.Create(&file).Error; err != nil {
		t.Fatalf("create managed media record: %v", err)
	}
	previewURL, err := createMediaPreviewURL(context.Background(), key, mediaPreviewURLTTL)
	if err != nil {
		t.Fatalf("create application preview URL: %v", err)
	}
	parsed, err := url.Parse(previewURL)
	if err != nil {
		t.Fatalf("parse application preview URL: %v", err)
	}
	token := parsed.Query().Get("token")

	resolved, err := (&FileUploadAndDownloadService{}).ResolveMediaPreviewObject(context.Background(), key, token)
	if err != nil || resolved.ID != file.ID {
		t.Fatalf("resolve managed preview: file=%#v err=%v", resolved, err)
	}
	if _, err = (&FileUploadAndDownloadService{}).ResolveMediaPreviewObject(context.Background(), key+"-tampered", token); !errors.Is(err, ErrInvalidMediaPreview) {
		t.Fatalf("tampered key should be rejected, got %v", err)
	}

	const externalKey = "assets/external.png"
	external := exampleModel.ExaFileUploadAndDownload{Name: "external.png", Key: externalKey, Url: "https://cdn.example.com/external.png"}
	if err = database.Create(&external).Error; err != nil {
		t.Fatalf("create external media record: %v", err)
	}
	externalPreviewURL, err := createMediaPreviewURL(context.Background(), externalKey, mediaPreviewURLTTL)
	if err != nil {
		t.Fatalf("create external test token: %v", err)
	}
	externalParsed, _ := url.Parse(externalPreviewURL)
	if _, err = (&FileUploadAndDownloadService{}).ResolveMediaPreviewObject(context.Background(), externalKey, externalParsed.Query().Get("token")); !errors.Is(err, ErrMediaPreviewMissing) {
		t.Fatalf("external media record should not resolve as private storage, got %v", err)
	}
}

func TestConfiguredMediaPreviewURLDoesNotRequireLibraryRecord(t *testing.T) {
	configureMediaPreviewTest(t)
	global.GVA_CONFIG.System.OssType = "minio"
	global.GVA_CONFIG.Minio.BucketUrl = "http://minio.internal/gva-assets"
	const key = "assets/brand/current-logo.png"

	previewURL, err := ResolveConfiguredMediaPreviewURL("http://minio.internal/gva-assets/" + key)
	if err != nil {
		t.Fatalf("resolve configured preview URL: %v", err)
	}
	parsed, err := url.Parse(previewURL)
	if err != nil {
		t.Fatalf("parse configured preview URL: %v", err)
	}
	resolved, err := (&FileUploadAndDownloadService{}).ResolveMediaPreviewObject(
		context.Background(), parsed.Query().Get("key"), parsed.Query().Get("token"),
	)
	if err != nil || resolved.Key != key {
		t.Fatalf("resolve configured media preview: file=%#v err=%v", resolved, err)
	}
}

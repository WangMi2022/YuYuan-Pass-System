package model

import (
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/global"
)

func configurePhotoTestMinio(t *testing.T) {
	t.Helper()
	previous := global.GVA_CONFIG.Minio
	t.Cleanup(func() { global.GVA_CONFIG.Minio = previous })
	global.GVA_CONFIG.Minio.BucketUrl = "http://172.30.3.135:9000/gva-assets"
	global.GVA_CONFIG.Minio.BasePath = "assets"
}

func TestResolvePhotoObjectKeyFromManagedLegacyURL(t *testing.T) {
	configurePhotoTestMinio(t)
	photo := Photo{URL: "http://172.30.3.135:9000/gva-assets/assets/2026-07-16/2773ab65220d8bc7fa931bd4c873af48.jpg"}

	key, ok := ResolvePhotoObjectKey(photo)
	if !ok || key != "assets/2026-07-16/2773ab65220d8bc7fa931bd4c873af48.jpg" {
		t.Fatalf("ResolvePhotoObjectKey() = %q, %t", key, ok)
	}
}

func TestResolvePhotoObjectKeyRejectsUnmanagedOrUnsafeURLs(t *testing.T) {
	configurePhotoTestMinio(t)
	tests := []string{
		"https://172.30.3.135:9000/gva-assets/assets/photo.jpg",
		"http://172.30.3.135:9001/gva-assets/assets/photo.jpg",
		"http://172.30.3.135:9000.evil.example/gva-assets/assets/photo.jpg",
		"http://172.30.3.135:9000/other-bucket/assets/photo.jpg",
		"http://172.30.3.135:9000/gva-assets/assets/../secret.jpg",
		"http://172.30.3.135:9000/gva-assets/uploads/photo.jpg",
		"http://172.30.3.135:9000/gva-assets/assets/photo.jpg?download=1",
		"https://cdn.example.com/assets/photo.jpg",
	}
	for _, rawURL := range tests {
		if key, ok := ResolvePhotoObjectKey(Photo{URL: rawURL}); ok {
			t.Errorf("ResolvePhotoObjectKey(%q) unexpectedly accepted %q", rawURL, key)
		}
	}
}

func TestNormalizeAssetPhotoBuildsAuthenticatedProxyURL(t *testing.T) {
	configurePhotoTestMinio(t)
	photo := Photo{
		URL:         "http://172.30.3.135:9000/gva-assets/assets/2026-07-16/photo.jpg",
		AccessToken: "expired-token",
	}

	NormalizeAssetPhoto(&photo, 42)

	if photo.Key != "assets/2026-07-16/photo.jpg" || photo.AssetID != 42 || photo.AccessToken != "" {
		t.Fatalf("NormalizeAssetPhoto() produced unexpected photo: %#v", photo)
	}
	if photo.URL != "/api/asset/photo?key=assets%2F2026-07-16%2Fphoto.jpg&assetId=42" {
		t.Fatalf("NormalizeAssetPhoto() URL = %q", photo.URL)
	}
}

func TestValidPhotoObjectKeySupportsNestedConfiguredPrefix(t *testing.T) {
	configurePhotoTestMinio(t)
	global.GVA_CONFIG.Minio.BasePath = "tenant/assets"

	if !ValidPhotoObjectKey("tenant/assets/2026-07-16/photo.jpg") {
		t.Fatal("nested configured base path should be accepted")
	}
	if ValidPhotoObjectKey("tenant/assets-other/photo.jpg") {
		t.Fatal("similar path outside configured base path must be rejected")
	}
}

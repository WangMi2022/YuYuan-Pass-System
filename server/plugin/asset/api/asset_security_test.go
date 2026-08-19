package api

import (
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/asset/model"
)

func TestIsAllowedAssetPhotoContentType(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{value: "image/jpeg", want: true},
		{value: "image/png; charset=binary", want: true},
		{value: "image/svg+xml", want: false},
		{value: "application/pdf", want: false},
		{value: "", want: false},
	}
	for _, test := range tests {
		if got := isAllowedAssetPhotoContentType(test.value); got != test.want {
			t.Fatalf("isAllowedAssetPhotoContentType(%q) = %t, want %t", test.value, got, test.want)
		}
	}
}

func TestAssetPhotoBelongsToAssetSupportsManagedLegacyURL(t *testing.T) {
	previous := global.GVA_CONFIG.Minio
	t.Cleanup(func() { global.GVA_CONFIG.Minio = previous })
	global.GVA_CONFIG.Minio.BucketUrl = "http://172.30.3.135:9000/gva-assets"
	global.GVA_CONFIG.Minio.BasePath = "assets"

	key := "assets/2026-07-16/2773ab65220d8bc7fa931bd4c873af48.jpg"
	photos := []model.Photo{{URL: "http://172.30.3.135:9000/gva-assets/" + key}}
	if !assetPhotoBelongsToAsset(photos, key) {
		t.Fatal("managed legacy photo URL should authorize its owning asset")
	}
	if assetPhotoBelongsToAsset([]model.Photo{{URL: "https://cdn.example.com/" + key}}, key) {
		t.Fatal("external photo URL must not authorize a private object key")
	}
}

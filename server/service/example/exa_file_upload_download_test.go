package example

import (
	"context"
	"errors"
	"testing"
	"time"

	exampleModel "github.com/WangMi2022/mit-assets-admin/server/model/example"
)

func TestAttachMediaPreviewURLsSignsManagedMinioObjectsWithoutReplacingCanonicalURLs(t *testing.T) {
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
		return "http://minio.local/gva-assets/" + key + "?X-Amz-Signature=test", nil
	}

	if err := attachMediaPreviewURLs(context.Background(), files, "minio", "http://minio.local/gva-assets/", signer); err != nil {
		t.Fatalf("attach media preview URLs: %v", err)
	}
	if len(signedKeys) != 1 || signedKeys[0] != files[0].Key {
		t.Fatalf("unexpected signed keys: %#v", signedKeys)
	}
	if files[0].PreviewURL != "http://minio.local/gva-assets/assets/2026-08-17/photo.jpg?X-Amz-Signature=test" {
		t.Fatalf("managed object did not receive a signed preview URL: %q", files[0].PreviewURL)
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
	files := []exampleModel.ExaFileUploadAndDownload{
		{Url: "http://minio.local/gva-assets/assets/broken.jpg", Key: "assets/broken.jpg"},
		{Url: "http://minio.local/gva-assets/assets/working.jpg", Key: "assets/working.jpg"},
	}
	signer := func(_ context.Context, key string, _ time.Duration) (string, error) {
		if key == "assets/broken.jpg" {
			return "", errors.New("test signing failure")
		}
		return "http://minio.local/gva-assets/" + key + "?X-Amz-Signature=test", nil
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
	signer := func(_ context.Context, key string, expires time.Duration) (string, error) {
		if key != "avatars/user-7.png" {
			t.Fatalf("unexpected avatar object key: %q", key)
		}
		if expires != mediaPreviewURLTTL {
			t.Fatalf("unexpected preview URL TTL: %s", expires)
		}
		return "http://minio.local/gva-assets/" + key + "?X-Amz-Signature=avatar", nil
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
	if previewURL != "http://minio.local/gva-assets/avatars/user-7.png?X-Amz-Signature=avatar" {
		t.Fatalf("private avatar did not receive a signed preview URL: %q", previewURL)
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

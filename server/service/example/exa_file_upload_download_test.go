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

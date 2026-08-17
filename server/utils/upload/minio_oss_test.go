package upload

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func TestMinioPreviewURLSignsTheRequestedObject(t *testing.T) {
	client, err := minio.New("minio.example.test", &minio.Options{
		Creds:  credentials.NewStaticV4("access-key", "secret-key", ""),
		Secure: true,
		Region: "us-east-1",
	})
	if err != nil {
		t.Fatalf("create MinIO client: %v", err)
	}
	storage := &Minio{Client: client, bucket: "gva-assets"}

	previewURL, err := storage.PreviewURL(context.Background(), "assets/photo.jpg", 15*time.Minute)
	if err != nil {
		t.Fatalf("create preview URL: %v", err)
	}
	parsed, err := url.Parse(previewURL)
	if err != nil {
		t.Fatalf("parse preview URL: %v", err)
	}
	if parsed.Host != "minio.example.test" || parsed.Path != "/gva-assets/assets/photo.jpg" {
		t.Fatalf("unexpected preview URL target: %s", previewURL)
	}
	if !strings.Contains(parsed.RawQuery, "X-Amz-Signature=") || parsed.Query().Get("X-Amz-Expires") != "900" {
		t.Fatalf("preview URL is not signed for 15 minutes: %s", previewURL)
	}
}

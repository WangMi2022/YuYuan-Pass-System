package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestCappedBufferLimitsCapturedBytes(t *testing.T) {
	buffer := newCappedBuffer(8)

	if n, err := buffer.Write([]byte("12345")); err != nil || n != 5 {
		t.Fatalf("first write = (%d, %v), want (5, nil)", n, err)
	}
	if n, err := buffer.Write([]byte("67890")); err != nil || n != 5 {
		t.Fatalf("second write = (%d, %v), want (5, nil)", n, err)
	}
	if got := buffer.Len(); got != 8 {
		t.Fatalf("captured length = %d, want 8", got)
	}
	if got := buffer.String(); got != "12345678" {
		t.Fatalf("captured body = %q, want %q", got, "12345678")
	}
	if got := buffer.Summary(); got != "12345678"+truncatedBodySummary {
		t.Fatalf("summary = %q, want captured prefix plus truncation marker", got)
	}
}

func TestCappedBufferSummaryIsValidUTF8(t *testing.T) {
	buffer := newCappedBuffer(2)
	_, _ = buffer.Write([]byte("中"))

	if summary := buffer.Summary(); !utf8.ValidString(summary) {
		t.Fatalf("summary contains invalid UTF-8: %q", summary)
	}
}

func TestTeeReadCloserPreservesFullRequestBody(t *testing.T) {
	original := strings.Repeat("request-body-", 200)
	capture := newCappedBuffer(32)
	body := io.NopCloser(strings.NewReader(original))
	wrapped := &teeReadCloser{Reader: io.TeeReader(body, capture), Closer: body}

	got, err := io.ReadAll(wrapped)
	if err != nil {
		t.Fatalf("read wrapped body: %v", err)
	}
	if !bytes.Equal(got, []byte(original)) {
		t.Fatal("wrapped body did not preserve the complete request")
	}
	if capture.Len() != 32 || !strings.HasSuffix(capture.Summary(), truncatedBodySummary) {
		t.Fatalf("capture = (%d, %q), want 32 captured bytes plus truncation marker", capture.Len(), capture.Summary())
	}
}

func TestIsBinaryResponse(t *testing.T) {
	tests := []struct {
		name   string
		header http.Header
		want   bool
	}{
		{name: "json", header: http.Header{"Content-Type": []string{"application/json"}}, want: false},
		{name: "attachment", header: http.Header{"Content-Disposition": []string{"attachment; filename=report.xlsx"}}, want: true},
		{name: "octet stream", header: http.Header{"Content-Type": []string{"application/octet-stream"}}, want: true},
		{name: "pdf", header: http.Header{"Content-Type": []string{"application/pdf"}}, want: true},
		{name: "xlsx", header: http.Header{"Content-Type": []string{"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"}}, want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isBinaryResponse(test.header); got != test.want {
				t.Fatalf("isBinaryResponse() = %v, want %v", got, test.want)
			}
		})
	}
}

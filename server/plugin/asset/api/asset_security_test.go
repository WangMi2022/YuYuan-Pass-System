package api

import "testing"

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

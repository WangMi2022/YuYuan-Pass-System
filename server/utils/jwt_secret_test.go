package utils

import "testing"

func TestIsWeakJWTSigningKey(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "empty", value: "", want: true},
		{name: "default", value: "qmPlus", want: true},
		{name: "short", value: "short-secret", want: true},
		{name: "strong", value: "12345678901234567890123456789012", want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsWeakJWTSigningKey(test.value); got != test.want {
				t.Fatalf("IsWeakJWTSigningKey(%q) = %t, want %t", test.value, got, test.want)
			}
		})
	}
}

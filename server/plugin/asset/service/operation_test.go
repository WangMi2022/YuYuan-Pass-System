package service

import "testing"

func TestTransitionStatus(t *testing.T) {
	tests := []struct {
		name, operationType, from, want string
		wantErr                         bool
	}{
		{name: "inbound pending asset", operationType: "inbound", from: "pending_inbound", want: "idle"},
		{name: "reject repeated inbound", operationType: "inbound", from: "idle", wantErr: true},
		{name: "issue idle asset", operationType: "issue", from: "idle", want: "in_use"},
		{name: "return used asset", operationType: "return", from: "in_use", want: "idle"},
		{name: "return repaired asset", operationType: "return", from: "maintenance", want: "idle"},
		{name: "maintenance used asset", operationType: "maintenance", from: "in_use", want: "maintenance"},
		{name: "scrap maintenance asset", operationType: "scrap", from: "maintenance", want: "retired"},
		{name: "reject retired asset", operationType: "issue", from: "retired", wantErr: true},
		{name: "reject invalid operation", operationType: "unknown", from: "idle", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transitionStatus(tt.operationType, tt.from)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got status %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

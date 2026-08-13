package middleware

import (
	"testing"
	"time"
)

func TestAllowAIRequestLimitsWindow(t *testing.T) {
	aiRateLimiter.Lock()
	aiRateLimiter.entries = make(map[string]aiRateEntry)
	aiRateLimiter.Unlock()

	now := time.Unix(100, 0)
	if !allowAIRequest("user:1", time.Minute, 2, now) {
		t.Fatal("first request should be allowed")
	}
	if !allowAIRequest("user:1", time.Minute, 2, now.Add(time.Second)) {
		t.Fatal("second request should be allowed")
	}
	if allowAIRequest("user:1", time.Minute, 2, now.Add(2*time.Second)) {
		t.Fatal("third request should be rate limited")
	}
	if !allowAIRequest("user:1", time.Minute, 2, now.Add(time.Minute)) {
		t.Fatal("request after window should be allowed")
	}
}

func TestAISSEConcurrencyReleasesAfterHandler(t *testing.T) {
	aiSSELimiter.Lock()
	aiSSELimiter.entries = make(map[string]int)
	aiSSELimiter.Unlock()

	if !acquireAISSE("user:1", 1) {
		t.Fatal("first SSE request should acquire a slot")
	}
	if acquireAISSE("user:1", 1) {
		t.Fatal("second SSE request should be rejected at the limit")
	}
	releaseAISSE("user:1")
	if !acquireAISSE("user:1", 1) {
		t.Fatal("slot should be available after release")
	}
	releaseAISSE("user:1")
}

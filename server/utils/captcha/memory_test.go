package captcha

import (
	"testing"
	"time"
)

func TestMemoryStoreExpiresChallenges(t *testing.T) {
	store := NewMemoryStore(time.Millisecond)
	if err := store.Set("challenge-id", "1234"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Millisecond)
	if store.Verify("challenge-id", "1234", true) {
		t.Fatal("expired memory captcha challenge was accepted")
	}
}

func TestMemoryStoreConsumesChallengeOnce(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	if err := store.Set("challenge-id", "1234"); err != nil {
		t.Fatal(err)
	}
	if !store.Verify("challenge-id", "1234", true) {
		t.Fatal("valid memory captcha challenge was rejected")
	}
	if store.Verify("challenge-id", "1234", true) {
		t.Fatal("memory captcha challenge was accepted more than once")
	}
}

func TestMemoryStoreRejectsEmptyID(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	if err := store.Set("   ", "1234"); err == nil {
		t.Fatal("empty memory captcha ID was accepted")
	}
}

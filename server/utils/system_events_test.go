package utils

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSystemEventsRejectsConcurrentReload(t *testing.T) {
	events := new(SystemEvents)
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	var releaseOnce sync.Once
	var active atomic.Int32
	var peak atomic.Int32
	var calls atomic.Int32
	releaseHandlers := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(releaseHandlers)

	events.RegisterReloadHandler(func() error {
		calls.Add(1)
		current := active.Add(1)
		defer active.Add(-1)
		for {
			previous := peak.Load()
			if current <= previous || peak.CompareAndSwap(previous, current) {
				break
			}
		}
		started <- struct{}{}
		<-release
		return nil
	})

	firstDone := make(chan error, 1)
	go func() { firstDone <- events.TriggerReload() }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first reload did not reach the handler")
	}

	secondDone := make(chan error, 1)
	go func() { secondDone <- events.TriggerReload() }()
	select {
	case err := <-secondDone:
		if err == nil || err.Error() != "系统配置正在重载，请稍后再试" {
			t.Fatalf("concurrent reload error = %v, want busy error", err)
		}
	case <-time.After(200 * time.Millisecond):
		releaseHandlers()
		t.Fatal("concurrent reload waited instead of being rejected")
	}

	releaseHandlers()
	select {
	case err := <-firstDone:
		if err != nil {
			t.Fatalf("first reload failed: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("first reload did not finish")
	}
	if calls.Load() != 1 || peak.Load() != 1 {
		t.Fatalf("reload calls = %d, peak concurrency = %d; want 1 and 1", calls.Load(), peak.Load())
	}
}

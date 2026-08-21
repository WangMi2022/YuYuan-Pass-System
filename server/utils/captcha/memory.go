package captcha

import (
	"errors"
	"strings"
	"sync"
	"time"
)

// MemoryStore is the single-process fallback for installations without Redis.
// Entries are checked against their expiry on every read, so an idle process
// cannot accept a challenge after its TTL has elapsed.
type MemoryStore struct {
	mu         sync.Mutex
	expiration time.Duration
	values     map[string]memoryEntry
}

type memoryEntry struct {
	value     string
	expiresAt time.Time
}

func NewMemoryStore(expiration time.Duration) *MemoryStore {
	if expiration <= 0 {
		expiration = DefaultExpiration
	}
	return &MemoryStore{
		expiration: expiration,
		values:     make(map[string]memoryEntry),
	}
}

func (s *MemoryStore) Set(id, value string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("captcha id is required")
	}
	now := time.Now()
	s.mu.Lock()
	s.removeExpiredLocked(now)
	s.values[id] = memoryEntry{value: value, expiresAt: now.Add(s.expiration)}
	s.mu.Unlock()
	return nil
}

func (s *MemoryStore) Get(id string, clear bool) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.values[id]
	if !ok {
		return ""
	}
	if !entry.expiresAt.After(now) {
		delete(s.values, id)
		return ""
	}
	if clear {
		delete(s.values, id)
	}
	return entry.value
}

func (s *MemoryStore) Verify(id, answer string, clear bool) bool {
	id = strings.TrimSpace(id)
	answer = strings.TrimSpace(answer)
	if id == "" || answer == "" {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(s.Get(id, clear)), answer)
}

func (s *MemoryStore) removeExpiredLocked(now time.Time) {
	for id, entry := range s.values {
		if !entry.expiresAt.After(now) {
			delete(s.values, id)
		}
	}
}

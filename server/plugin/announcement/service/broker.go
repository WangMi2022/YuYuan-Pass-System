package service

import "sync"

type AnnouncementEvent struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
}

type announcementBroker struct {
	mu      sync.RWMutex
	clients map[chan AnnouncementEvent]struct{}
}

var NotificationHub = &announcementBroker{clients: make(map[chan AnnouncementEvent]struct{})}

func (b *announcementBroker) Subscribe() chan AnnouncementEvent {
	ch := make(chan AnnouncementEvent, 4)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *announcementBroker) Unsubscribe(ch chan AnnouncementEvent) {
	b.mu.Lock()
	if _, ok := b.clients[ch]; ok {
		delete(b.clients, ch)
		close(ch)
	}
	b.mu.Unlock()
}

func (b *announcementBroker) Publish(event AnnouncementEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- event:
		default:
		}
	}
}

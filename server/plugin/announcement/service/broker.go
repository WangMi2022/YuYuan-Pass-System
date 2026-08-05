package service

import "sync"

type NotificationEvent struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
	Kind  string `json:"kind"`
}

// AnnouncementEvent is retained for existing announcement publishers.
type AnnouncementEvent = NotificationEvent

type announcementBroker struct {
	mu      sync.RWMutex
	clients map[chan NotificationEvent]uint
}

var NotificationHub = &announcementBroker{clients: make(map[chan NotificationEvent]uint)}

func (b *announcementBroker) Subscribe(userID uint) chan NotificationEvent {
	ch := make(chan NotificationEvent, 4)
	b.mu.Lock()
	b.clients[ch] = userID
	b.mu.Unlock()
	return ch
}

func (b *announcementBroker) Unsubscribe(ch chan NotificationEvent) {
	b.mu.Lock()
	if _, ok := b.clients[ch]; ok {
		delete(b.clients, ch)
		close(ch)
	}
	b.mu.Unlock()
}

func (b *announcementBroker) Publish(event AnnouncementEvent) {
	b.publish(event, 0)
}

func (b *announcementBroker) PublishToUser(userID uint, event NotificationEvent) {
	if userID == 0 {
		return
	}
	b.publish(event, userID)
}

func (b *announcementBroker) publish(event NotificationEvent, targetUserID uint) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch, userID := range b.clients {
		if targetUserID != 0 && userID != targetUserID {
			continue
		}
		select {
		case ch <- event:
		default:
		}
	}
}

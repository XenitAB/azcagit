package cache

import (
	"context"

	"github.com/xenitab/azcagit/src/notification"
)

type InMemNotificationCache struct {
	event *notification.NotificationEvent
}

var _ NotificationCache = (*InMemNotificationCache)(nil)

func NewInMemNotificationCache() *InMemNotificationCache {
	return &InMemNotificationCache{}
}

func (c *InMemNotificationCache) Set(ctx context.Context, event notification.NotificationEvent) error {
	c.event = &event
	return nil
}

func (c *InMemNotificationCache) Get(ctx context.Context) (notification.NotificationEvent, bool, error) {
	if c.event == nil {
		return notification.NotificationEvent{}, false, nil
	}

	return *c.event, true, nil
}

func (c *InMemNotificationCache) Reset() {
	c.event = nil
}

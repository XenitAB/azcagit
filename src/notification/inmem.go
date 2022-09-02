package notification

import "context"

type InMemNotification struct{}

var _ Notification = (*InMemNotification)(nil)

func NewInMemNotification() *InMemNotification {
	return &InMemNotification{}
}

func (n *InMemNotification) Send(ctx context.Context, event NotificationEvent) error {
	return nil
}

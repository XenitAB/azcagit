package notification

import "context"

type DiscardNotification struct{}

var _ Notification = (*DiscardNotification)(nil)

func NewDiscardNotification() *DiscardNotification {
	return &DiscardNotification{}
}

func (n *DiscardNotification) Send(_ context.Context, _ NotificationEvent) error {
	return nil
}

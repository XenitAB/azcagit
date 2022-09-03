package notification

import "context"

type InMemNotification struct {
	notifications []NotificationEvent
	sendResponse  struct {
		err error
	}
}

var _ Notification = (*InMemNotification)(nil)

func NewInMemNotification() *InMemNotification {
	return &InMemNotification{
		notifications: []NotificationEvent{},
	}
}

func (n *InMemNotification) Send(ctx context.Context, event NotificationEvent) error {
	n.notifications = append(n.notifications, event)
	return n.sendResponse.err
}

func (n *InMemNotification) SendResponse(err error) {
	n.sendResponse.err = err
}

func (n *InMemNotification) ResetNotifications() {
	n.notifications = []NotificationEvent{}
}

func (n *InMemNotification) GetNotifications() []NotificationEvent {
	return n.notifications
}

package cache

import (
	"context"

	"github.com/xenitab/azcagit/src/azure"
	"github.com/xenitab/azcagit/src/config"
	"github.com/xenitab/azcagit/src/notification"
)

type CosmosDBNotificationCache struct {
	client *azure.CosmosDBContainerClient[notification.NotificationEvent]
}

var _ NotificationCache = (*CosmosDBNotificationCache)(nil)

const notificationCacheKey = "previous_notification"

func NewCosmosDBNotificationCache(cfg config.ReconcileConfig, cosmosDBClient *azure.CosmosDBClient) (*CosmosDBNotificationCache, error) {
	ttl := -1 // -1 disables time to live
	client, err := azure.NewCosmosDBContainerClient[notification.NotificationEvent](cosmosDBClient, "notification-cache", &ttl)
	if err != nil {
		return nil, err
	}

	return &CosmosDBNotificationCache{
		client,
	}, nil
}

func (c *CosmosDBNotificationCache) Set(ctx context.Context, event notification.NotificationEvent) error {
	return c.client.Set(ctx, notificationCacheKey, event)
}

func (c *CosmosDBNotificationCache) Get(ctx context.Context) (notification.NotificationEvent, bool, error) {
	value, err := c.client.Get(ctx, notificationCacheKey)
	if err != nil {
		return notification.NotificationEvent{}, false, err
	}

	if value == nil {
		return notification.NotificationEvent{}, false, nil
	}

	return *value, true, nil
}

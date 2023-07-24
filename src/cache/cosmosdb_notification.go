package cache

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/xenitab/azcagit/src/azure"
	"github.com/xenitab/azcagit/src/config"
	"github.com/xenitab/azcagit/src/notification"
)

type cosmosDBNotificationEventEntry struct {
	Id    string                         `json:"id"`
	Name  string                         `json:"name"`
	Event notification.NotificationEvent `json:"event"`
}

type CosmosDBNotificationCache struct {
	client *azure.CosmosDBClient[cosmosDBNotificationEventEntry]
}

var _ NotificationCache = (*CosmosDBNotificationCache)(nil)

func NewCosmosDBNotificationCache(cfg config.ReconcileConfig, cred azcore.TokenCredential) (*CosmosDBNotificationCache, error) {
	client, err := azure.NewCosmosDBClient[cosmosDBNotificationEventEntry](cfg.CosmosDBAccount, cfg.CosmosDBSqlDb, cfg.CosmosDBNotificationCacheContainer, cred)
	if err != nil {
		return nil, err
	}

	return &CosmosDBNotificationCache{
		client,
	}, nil
}

func (c *CosmosDBNotificationCache) Set(ctx context.Context, event notification.NotificationEvent) error {
	cacheEntry := cosmosDBNotificationEventEntry{
		Id:    "previous_notification",
		Name:  "previous_notification",
		Event: event,
	}

	return c.client.Set(ctx, cacheEntry.Name, cacheEntry)
}

func (c *CosmosDBNotificationCache) Get(ctx context.Context) (notification.NotificationEvent, bool, error) {
	cacheEntry, err := c.client.Get(ctx, "previous_notification")
	if err != nil {
		return notification.NotificationEvent{}, false, err
	}

	if cacheEntry == nil {
		return notification.NotificationEvent{}, false, nil
	}

	return cacheEntry.Event, true, nil
}

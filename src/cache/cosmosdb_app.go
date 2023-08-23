package cache

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
	"github.com/xenitab/azcagit/src/azure"
	"github.com/xenitab/azcagit/src/config"
)

type CosmosDBAppCache struct {
	client *azure.CosmosDBContainerClient[CacheEntry]
}

var _ AppCache = (*CosmosDBAppCache)(nil)

func NewCosmosDBAppCache(cfg config.ReconcileConfig, cosmosDBClient *azure.CosmosDBClient) (*CosmosDBAppCache, error) {
	ttl := 3600
	client, err := azure.NewCosmosDBContainerClient[CacheEntry](cosmosDBClient, "app-cache", &ttl)
	if err != nil {
		return nil, err
	}

	return &CosmosDBAppCache{
		client,
	}, nil
}

func (c *CosmosDBAppCache) Set(ctx context.Context, name string, remoteApp, sourceApp *armappcontainers.ContainerApp) error {
	if remoteApp == nil {
		return nil
	}
	if remoteApp.SystemData == nil {
		return nil
	}

	timestamp := remoteApp.SystemData.LastModifiedAt
	if timestamp == nil {
		if remoteApp.SystemData.CreatedAt == nil {
			return nil
		}
		timestamp = remoteApp.SystemData.CreatedAt
	}

	b, err := sourceApp.MarshalJSON()
	if err != nil {
		return nil
	}

	hash := fmt.Sprintf("%x", md5.Sum(b))
	cacheEntry := newCacheEntry(name, *timestamp, hash)
	return c.client.Set(ctx, name, cacheEntry)
}

func (c *CosmosDBAppCache) NeedsUpdate(ctx context.Context, name string, remoteApp, sourceApp *armappcontainers.ContainerApp) (bool, string, error) {
	entry, err := c.client.Get(ctx, name)
	if err != nil {
		return false, "CosmosDB client returned an error", err
	}

	if entry == nil {
		return true, "not in AppCache", nil
	}

	if remoteApp == nil {
		return true, "remoteApp nil", nil
	}
	if remoteApp.SystemData == nil {
		return true, "remoteApp SystemData nil", nil
	}

	if remoteApp.SystemData.LastModifiedAt != nil {
		if entry.Modified.Round(time.Millisecond) != (*remoteApp.SystemData.LastModifiedAt).Round(time.Millisecond) {
			return true, "changed LastModifiedAt", nil
		}
	} else if remoteApp.SystemData.CreatedAt != nil {
		if entry.Modified.Round(time.Millisecond) != (*remoteApp.SystemData.CreatedAt).Round(time.Millisecond) {
			return true, "changed CreatedAt", nil
		}
	}

	b, err := sourceApp.MarshalJSON()
	if err != nil {
		return true, "sourceApp MarshalJSON() failed", nil
	}

	hash := fmt.Sprintf("%x", md5.Sum(b))

	if entry.Hash != hash {
		return true, "changed sourceApp hash", nil
	}

	return false, "no changes", nil
}

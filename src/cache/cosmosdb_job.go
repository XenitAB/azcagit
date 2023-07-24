package cache

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
	"github.com/xenitab/azcagit/src/azure"
	"github.com/xenitab/azcagit/src/config"
)

type CosmosDBJobCache struct {
	client *azure.CosmosDBClient[CacheEntry]
}

var _ JobCache = (*CosmosDBJobCache)(nil)

func NewCosmosDBJobCache(cfg config.ReconcileConfig, cred azcore.TokenCredential) (*CosmosDBJobCache, error) {
	client, err := azure.NewCosmosDBClient[CacheEntry](cfg.CosmosDBAccount, cfg.CosmosDBSqlDb, cfg.CosmosDBJobCacheContainer, cred)
	if err != nil {
		return nil, err
	}

	return &CosmosDBJobCache{
		client,
	}, nil
}

func (c *CosmosDBJobCache) Set(ctx context.Context, name string, remoteJob, sourceJob *armappcontainers.Job) error {
	if remoteJob == nil {
		return nil
	}
	if remoteJob.SystemData == nil {
		return nil
	}

	timestamp := remoteJob.SystemData.LastModifiedAt
	if timestamp == nil {
		if remoteJob.SystemData.CreatedAt == nil {
			return nil
		}
		timestamp = remoteJob.SystemData.CreatedAt
	}

	b, err := sourceJob.MarshalJSON()
	if err != nil {
		return nil
	}

	hash := fmt.Sprintf("%x", md5.Sum(b))
	cacheEntry := newCacheEntry(name, *timestamp, hash)
	return c.client.Set(ctx, name, cacheEntry)
}

func (c *CosmosDBJobCache) NeedsUpdate(ctx context.Context, name string, remoteJob, sourceJob *armappcontainers.Job) (bool, string, error) {
	entry, err := c.client.Get(ctx, name)
	if err != nil {
		return false, "CosmosDB client returned an error", err
	}

	if entry == nil {
		return true, "not in JobCache", nil
	}

	if remoteJob == nil {
		return true, "remoteJob nil", nil
	}
	if remoteJob.SystemData == nil {
		return true, "remoteJob SystemData nil", nil
	}

	if remoteJob.SystemData.LastModifiedAt != nil {
		if (entry.Modified).Round(time.Millisecond) != (*remoteJob.SystemData.LastModifiedAt).Round(time.Millisecond) {
			return true, "changed LastModifiedAt", nil
		}
	} else if remoteJob.SystemData.CreatedAt != nil {
		if entry.Modified.Round(time.Millisecond) != (*remoteJob.SystemData.CreatedAt).Round(time.Millisecond) {
			return true, "changed CreatedAt", nil
		}
	}

	b, err := sourceJob.MarshalJSON()
	if err != nil {
		return true, "remoteJob MarshalJSON() failed", nil
	}

	hash := fmt.Sprintf("%x", md5.Sum(b))

	if entry.Hash != hash {
		return true, "changed remoteJob hash", nil
	}

	return false, "no changes", nil
}

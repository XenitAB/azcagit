package cache

import (
	"context"

	"github.com/xenitab/azcagit/src/azure"
	"github.com/xenitab/azcagit/src/config"
)

type CosmosDBRevisionCache struct {
	client *azure.CosmosDBContainerClient[string]
}

var _ RevisionCache = (*CosmosDBRevisionCache)(nil)

const revisionCacheKey = "revision"

func NewCosmosDBRevisionCache(cfg config.ReconcileConfig, cosmosDBClient *azure.CosmosDBClient) (*CosmosDBRevisionCache, error) {
	ttl := -1 // -1 disables time to live
	client, err := azure.NewCosmosDBContainerClient[string](cosmosDBClient, "revision-cache", &ttl)
	if err != nil {
		return nil, err
	}

	return &CosmosDBRevisionCache{
		client,
	}, nil
}

func (c *CosmosDBRevisionCache) Set(ctx context.Context, revision string) error {
	return c.client.Set(ctx, revisionCacheKey, revision)
}

func (c *CosmosDBRevisionCache) Get(ctx context.Context) (string, error) {
	value, err := c.client.Get(ctx, revisionCacheKey)
	if err != nil {
		return "", err
	}

	if value == nil {
		return "", nil
	}

	return *value, nil
}

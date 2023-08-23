package cache

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

type InMemAppCache map[string]CacheEntry

var _ AppCache = (*InMemAppCache)(nil)

func NewInMemAppCache() *InMemAppCache {
	c := make(InMemAppCache)
	return &c
}

func (c *InMemAppCache) Set(ctx context.Context, name string, remoteApp, sourceApp *armappcontainers.ContainerApp) error {
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

	(*c)[name] = newCacheEntry(name, *timestamp, hash)

	return nil
}

func (c *InMemAppCache) NeedsUpdate(ctx context.Context, name string, remoteApp, sourceApp *armappcontainers.ContainerApp) (bool, string, error) {
	entry, ok := (*c)[name]
	if !ok {
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

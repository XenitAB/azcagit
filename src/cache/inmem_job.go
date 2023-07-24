package cache

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

type InMemJobCache map[string]CacheEntry

var _ JobCache = (*InMemJobCache)(nil)

func NewInMemJobCache() *InMemJobCache {
	c := make(InMemJobCache)
	return &c
}

func (c *InMemJobCache) Set(ctx context.Context, name string, remoteJob, sourceJob *armappcontainers.Job) error {
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

	(*c)[name] = newCacheEntry(name, *timestamp, hash)

	return nil
}

func (c *InMemJobCache) NeedsUpdate(ctx context.Context, name string, remoteJob, sourceJob *armappcontainers.Job) (bool, string, error) {
	entry, ok := (*c)[name]
	if !ok {
		return true, "not in JobCache", nil
	}

	if remoteJob == nil {
		return true, "remoteJob nil", nil
	}
	if remoteJob.SystemData == nil {
		return true, "remoteJob SystemData nil", nil
	}

	if remoteJob.SystemData.LastModifiedAt != nil {
		if entry.Modified.Round(time.Millisecond) != (*remoteJob.SystemData.LastModifiedAt).Round(time.Millisecond) {
			return true, "changed LastModifiedAt", nil
		}
	} else if remoteJob.SystemData.CreatedAt != nil {
		if entry.Modified.Round(time.Millisecond) != (*remoteJob.SystemData.CreatedAt).Round(time.Millisecond) {
			return true, "changed CreatedAt", nil
		}
	}

	b, err := sourceJob.MarshalJSON()
	if err != nil {
		return true, "sourceJob MarshalJSON() failed", nil
	}

	hash := fmt.Sprintf("%x", md5.Sum(b))

	if entry.Hash != hash {
		return true, "changed sourceJob hash", nil
	}

	return false, "no changes", nil
}

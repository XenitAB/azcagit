package cache

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
)

type CacheEntry struct {
	modified  time.Time
	localHash string
}

type Cache map[string]CacheEntry

func NewCache() *Cache {
	c := make(Cache)
	return &c
}

func (c *Cache) Set(name string, live, local *armappcontainers.ContainerApp) {
	if live == nil {
		return
	}
	if live.SystemData == nil {
		return
	}

	timestamp := live.SystemData.LastModifiedAt
	if timestamp == nil {
		if live.SystemData.CreatedAt == nil {
			return
		}
		timestamp = live.SystemData.CreatedAt
	}

	b, err := local.MarshalJSON()
	if err != nil {
		return
	}
	hash := fmt.Sprintf("%x", md5.Sum(b))

	(*c)[name] = CacheEntry{
		modified:  *timestamp,
		localHash: hash,
	}
}

func (c *Cache) NeedsUpdate(name string, live, local *armappcontainers.ContainerApp) bool {
	entry, ok := (*c)[name]
	if !ok {
		return true
	}

	if live == nil {
		return true
	}
	if live.SystemData == nil {
		return true
	}

	timestamp := live.SystemData.LastModifiedAt
	if timestamp == nil {
		if live.SystemData.CreatedAt == nil {
			return true
		}
		timestamp = live.SystemData.CreatedAt
	}

	if entry.modified != *timestamp {
		return true
	}

	b, err := local.MarshalJSON()
	if err != nil {
		return true
	}

	hash := fmt.Sprintf("%x", md5.Sum(b))
	return entry.localHash != hash
}

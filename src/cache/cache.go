package cache

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
)

type CacheEntry struct {
	modified      time.Time
	sourceAppHash string
}

type Cache map[string]CacheEntry

func NewCache() *Cache {
	c := make(Cache)
	return &c
}

func (c *Cache) Set(name string, remoteApp, sourceApp *armappcontainers.ContainerApp) {
	if remoteApp == nil {
		return
	}
	if remoteApp.SystemData == nil {
		return
	}

	timestamp := remoteApp.SystemData.LastModifiedAt
	if timestamp == nil {
		if remoteApp.SystemData.CreatedAt == nil {
			return
		}
		timestamp = remoteApp.SystemData.CreatedAt
	}

	b, err := sourceApp.MarshalJSON()
	if err != nil {
		return
	}
	hash := fmt.Sprintf("%x", md5.Sum(b))

	(*c)[name] = CacheEntry{
		modified:      *timestamp,
		sourceAppHash: hash,
	}
}

func (c *Cache) NeedsUpdate(name string, remoteApp, sourceApp *armappcontainers.ContainerApp) bool {
	entry, ok := (*c)[name]
	if !ok {
		return true
	}

	if remoteApp == nil {
		return true
	}
	if remoteApp.SystemData == nil {
		return true
	}

	timestamp := remoteApp.SystemData.LastModifiedAt
	if timestamp == nil {
		if remoteApp.SystemData.CreatedAt == nil {
			return true
		}
		timestamp = remoteApp.SystemData.CreatedAt
	}

	if entry.modified != *timestamp {
		return true
	}

	b, err := sourceApp.MarshalJSON()
	if err != nil {
		return true
	}

	hash := fmt.Sprintf("%x", md5.Sum(b))
	return entry.sourceAppHash != hash
}

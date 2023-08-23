package cache

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
	"github.com/xenitab/azcagit/src/notification"
)

type CacheEntry struct {
	Name     string    `json:"name"`
	Modified time.Time `json:"modified"`
	Hash     string    `json:"hash"`
}

func newCacheEntry(name string, modified time.Time, hash string) CacheEntry {
	return CacheEntry{
		Name:     name,
		Modified: modified.Round(time.Millisecond),
		Hash:     hash,
	}
}

type AppCache interface {
	Set(ctx context.Context, name string, remoteApp, sourceApp *armappcontainers.ContainerApp) error
	NeedsUpdate(ctx context.Context, name string, remoteApp, sourceApp *armappcontainers.ContainerApp) (bool, string, error)
}

type JobCache interface {
	Set(ctx context.Context, name string, remoteJob, sourceJob *armappcontainers.Job) error
	NeedsUpdate(ctx context.Context, name string, remoteJob, sourceJob *armappcontainers.Job) (bool, string, error)
}

type SecretCacheEntry struct {
	name     string
	value    string
	modified time.Time
}

type NotificationCache interface {
	Set(ctx context.Context, event notification.NotificationEvent) error
	Get(ctx context.Context) (notification.NotificationEvent, bool, error)
}

type RevisionCache interface {
	Set(ctx context.Context, revision string) error
	Get(ctx context.Context) (string, error)
}

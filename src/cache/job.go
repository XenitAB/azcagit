package cache

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

type JobCacheEntry struct {
	modified      time.Time
	sourceJobHash string
}

type JobCache map[string]JobCacheEntry

func NewJobCache() *JobCache {
	c := make(JobCache)
	return &c
}

func (c *JobCache) Set(name string, remoteJob, sourceJob *armappcontainers.Job) {
	if remoteJob == nil {
		return
	}
	if remoteJob.SystemData == nil {
		return
	}

	timestamp := remoteJob.SystemData.LastModifiedAt
	if timestamp == nil {
		if remoteJob.SystemData.CreatedAt == nil {
			return
		}
		timestamp = remoteJob.SystemData.CreatedAt
	}

	b, err := sourceJob.MarshalJSON()
	if err != nil {
		return
	}
	hash := fmt.Sprintf("%x", md5.Sum(b))

	(*c)[name] = JobCacheEntry{
		modified:      *timestamp,
		sourceJobHash: hash,
	}
}

func (c *JobCache) NeedsUpdate(name string, remoteJob, sourceJob *armappcontainers.Job) (bool, string) {
	entry, ok := (*c)[name]
	if !ok {
		return true, "not in JobCache"
	}

	if remoteJob == nil {
		return true, "remoteJob nil"
	}
	if remoteJob.SystemData == nil {
		return true, "remoteJob SystemData nil"
	}

	if remoteJob.SystemData.LastModifiedAt != nil {
		if entry.modified != *remoteJob.SystemData.LastModifiedAt {
			return true, "changed LastModifiedAt"
		}
	} else if remoteJob.SystemData.CreatedAt != nil {
		if entry.modified != *remoteJob.SystemData.CreatedAt {
			return true, "changed CreatedAt"
		}
	}

	b, err := sourceJob.MarshalJSON()
	if err != nil {
		return true, "sourceJob MarshalJSON() failed"
	}

	hash := fmt.Sprintf("%x", md5.Sum(b))

	if entry.sourceJobHash != hash {
		return true, "changed sourceJob hash"
	}

	return false, "no changes"
}

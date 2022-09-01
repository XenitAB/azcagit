package cache

import "time"

type SecretCacheEntry struct {
	name     string
	value    string
	modified time.Time
}
type SecretCache map[string]SecretCacheEntry

func NewSecretCache() *SecretCache {
	c := make(SecretCache)
	return &c
}

func (c *SecretCache) Set(name string, value string, modified time.Time) {
	(*c)[name] = SecretCacheEntry{
		name,
		value,
		modified,
	}
}

func (c *SecretCache) Get(name string) (string, bool) {
	entry, ok := (*c)[name]
	return entry.value, ok
}

func (c *SecretCache) NeedsUpdate(name string, modified time.Time) bool {
	entry, ok := (*c)[name]
	if !ok {
		return true
	}

	if modified != entry.modified {
		return true
	}

	return false
}

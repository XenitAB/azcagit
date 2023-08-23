package cache

import "time"

type InMemSecretCache map[string]SecretCacheEntry

func NewInMemSecretCache() *InMemSecretCache {
	c := make(InMemSecretCache)
	return &c
}

func (c *InMemSecretCache) Set(name string, value string, modified time.Time) {
	(*c)[name] = SecretCacheEntry{
		name,
		value,
		modified,
	}
}

func (c *InMemSecretCache) Get(name string) (string, bool) {
	entry, ok := (*c)[name]
	return entry.value, ok
}

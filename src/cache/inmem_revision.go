package cache

import (
	"context"
)

type InMemRevisionCache struct {
	revision *string
}

var _ RevisionCache = (*InMemRevisionCache)(nil)

func NewInMemRevisionCache() *InMemRevisionCache {
	return &InMemRevisionCache{}
}

func (c *InMemRevisionCache) Set(ctx context.Context, revision string) error {
	c.revision = &revision
	return nil
}

func (c *InMemRevisionCache) Get(ctx context.Context) (string, error) {
	if c.revision == nil {
		return "", nil
	}

	return *c.revision, nil
}

func (c *InMemRevisionCache) Reset() {
	c.revision = nil
}

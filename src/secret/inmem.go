package secret

import (
	"context"
	"fmt"
	"time"
)

type InMemSecret struct {
	items *Items
}

var _ Secret = (*InMemSecret)(nil)

func NewInMemSecret() *InMemSecret {
	items := make(Items)
	return &InMemSecret{
		items: &items,
	}
}

func (s *InMemSecret) ListItems(ctx context.Context) (*Items, error) {
	return s.items, nil
}

func (s *InMemSecret) Get(ctx context.Context, name string) (string, time.Time, error) {
	item, ok := s.items.Get(name)
	if !ok {
		return "", time.Time{}, fmt.Errorf("item %q not found", name)
	}
	return item.name, item.changedAt, nil
}

func (s *InMemSecret) Set(items *Items) {
	s.items = items
}

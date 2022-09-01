package secret

import (
	"context"
	"fmt"
	"time"
)

type InMemSecret struct {
	items  *Items
	values map[string]string
}

var _ Secret = (*InMemSecret)(nil)

func NewInMemSecret() *InMemSecret {
	items := make(Items)
	values := make(map[string]string)
	return &InMemSecret{
		items:  &items,
		values: values,
	}
}

func (s *InMemSecret) ListItems(ctx context.Context) (*Items, error) {
	return s.items, nil
}

func (s *InMemSecret) Get(ctx context.Context, name string) (string, time.Time, error) {
	item, ok := s.items.Get(name)
	if !ok {
		return "", time.Time{}, fmt.Errorf("item for %q not found", name)
	}

	value, ok := s.values[name]
	if !ok {
		return "", time.Time{}, fmt.Errorf("value for %q not found", name)
	}

	return value, item.changedAt, nil
}

func (s *InMemSecret) Reset() {
	items := make(Items)
	values := make(map[string]string)
	s.items = &items
	s.values = values
}

func (s *InMemSecret) Set(name string, value string, changedAt time.Time) {
	s.values[name] = value
	(*s.items)[name] = Item{
		name,
		changedAt,
	}
}

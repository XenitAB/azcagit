package secret

import (
	"time"
)

type Item struct {
	name      string
	changedAt time.Time
}

func (i *Item) LastChange() time.Time {
	return i.changedAt
}

func (i *Item) Name() string {
	return i.name
}

type Items map[string]Item

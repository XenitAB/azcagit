package secret

import (
	"context"
	"time"
)

type Secret interface {
	ListItems(ctx context.Context) (*Items, error)
	Get(ctx context.Context, name string) (string, time.Time, error)
}

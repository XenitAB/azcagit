package secret

import "context"

type Secret interface {
	ListItems(ctx context.Context) (*Items, error)
	Get(ctx context.Context, name string) (string, error)
}

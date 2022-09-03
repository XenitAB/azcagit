package source

import (
	"context"
)

type Source interface {
	Get(ctx context.Context) (*SourceApps, string, error)
}

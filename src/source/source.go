package source

import (
	"context"
)

type Sources struct {
	Apps *SourceApps
	Jobs *SourceJobs
}

type Source interface {
	Get(ctx context.Context) (*Sources, string, error)
}

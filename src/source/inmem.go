package source

import (
	"context"

	"github.com/xenitab/aca-gitops-engine/src/config"
)

type InMemSource struct{}

var _ Source = (*InMemSource)(nil)

func (s *InMemSource) Get(ctx context.Context) (*SourceApps, error) {
	return nil, nil
}

func NewInMemSource(cfg config.Config) (*InMemSource, error) {
	return nil, nil
}

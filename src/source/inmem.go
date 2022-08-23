package source

import (
	"context"

	"github.com/xenitab/aca-gitops-engine/src/config"
)

type InMemSource struct {
	getResponse struct {
		sourceApps *SourceApps
		err        error
	}
}

var _ Source = (*InMemSource)(nil)

func NewInMemSource(cfg config.Config) (*InMemSource, error) {
	return &InMemSource{}, nil
}

func (s *InMemSource) Get(ctx context.Context) (*SourceApps, error) {
	return s.getResponse.sourceApps, s.getResponse.err
}

func (s *InMemSource) GetResponse(sourceApps *SourceApps, err error) {
	s.getResponse.sourceApps = sourceApps
	s.getResponse.err = err
}

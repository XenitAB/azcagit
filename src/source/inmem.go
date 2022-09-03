package source

import (
	"context"
)

type InMemSource struct {
	getResponse struct {
		sourceApps *SourceApps
		revision   string
		err        error
	}
}

var _ Source = (*InMemSource)(nil)

func NewInMemSource() *InMemSource {
	return &InMemSource{}
}

func (s *InMemSource) Get(ctx context.Context) (*SourceApps, string, error) {
	return s.getResponse.sourceApps, s.getResponse.revision, s.getResponse.err
}

func (s *InMemSource) GetResponse(sourceApps *SourceApps, revision string, err error) {
	s.getResponse.sourceApps = sourceApps
	s.getResponse.revision = revision
	s.getResponse.err = err
}

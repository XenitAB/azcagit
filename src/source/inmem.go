package source

import (
	"context"
)

type InMemSource struct {
	getResponse struct {
		sourceApps *SourceApps
		err        error
	}
}

var _ Source = (*InMemSource)(nil)

func NewInMemSource() *InMemSource {
	return &InMemSource{}
}

func (s *InMemSource) Get(ctx context.Context) (*SourceApps, error) {
	return s.getResponse.sourceApps, s.getResponse.err
}

func (s *InMemSource) GetResponse(sourceApps *SourceApps, err error) {
	s.getResponse.sourceApps = sourceApps
	s.getResponse.err = err
}

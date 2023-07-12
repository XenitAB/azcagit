package source

import (
	"context"
)

type InMemSource struct {
	getResponse struct {
		sources  *Sources
		revision string
		err      error
	}
}

var _ Source = (*InMemSource)(nil)

func NewInMemSource() *InMemSource {
	return &InMemSource{}
}

func (s *InMemSource) Get(ctx context.Context) (*Sources, string, error) {
	return s.getResponse.sources, s.getResponse.revision, s.getResponse.err
}

func (s *InMemSource) GetResponse(sources *Sources, revision string, err error) {
	s.getResponse.sources = sources
	s.getResponse.revision = revision
	s.getResponse.err = err
}

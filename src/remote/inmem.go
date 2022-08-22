package remote

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
	"github.com/xenitab/aca-gitops-engine/src/config"
)

type InMemRemote struct{}

var _ Remote = (*InMemRemote)(nil)

func NewInMemRemote(cfg config.Config) (*InMemRemote, error) {
	return nil, nil
}

func (r *InMemRemote) List(ctx context.Context) (*RemoteApps, error) {
	return nil, nil
}

func (r *InMemRemote) CreateOrUpdate(ctx context.Context, name string, app armappcontainers.ContainerApp) error {
	return nil
}

func (r *InMemRemote) Delete(ctx context.Context, name string) error {
	return nil
}

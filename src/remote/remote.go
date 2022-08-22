package remote

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
)

type Remote interface {
	List(ctx context.Context) (*RemoteApps, error)
	CreateOrUpdate(ctx context.Context, name string, app armappcontainers.ContainerApp) error
	Delete(ctx context.Context, name string) error
}

package remote

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

type Remote interface {
	Get(ctx context.Context) (*RemoteApps, error)
	Create(ctx context.Context, name string, app armappcontainers.ContainerApp) error
	Update(ctx context.Context, name string, app armappcontainers.ContainerApp) error
	Delete(ctx context.Context, name string) error
}

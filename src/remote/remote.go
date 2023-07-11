package remote

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

type App interface {
	Get(ctx context.Context) (*RemoteApps, error)
	Create(ctx context.Context, name string, app armappcontainers.ContainerApp) error
	Update(ctx context.Context, name string, app armappcontainers.ContainerApp) error
	Delete(ctx context.Context, name string) error
}

type Job interface {
	Get(ctx context.Context) (*RemoteJobs, error)
	Create(ctx context.Context, name string, app armappcontainers.Job) error
	Update(ctx context.Context, name string, app armappcontainers.Job) error
	Delete(ctx context.Context, name string) error
}

package remote

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
	"github.com/xenitab/azcagit/src/config"
)

type AzureRemote struct {
	resourceGroup string
	client        *armappcontainers.ContainerAppsClient
}

var _ Remote = (*AzureRemote)(nil)

func NewAzureRemote(cfg config.Config, cred azcore.TokenCredential) (*AzureRemote, error) {
	client, err := armappcontainers.NewContainerAppsClient(cfg.SubscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	return &AzureRemote{
		resourceGroup: cfg.ResourceGroupName,
		client:        client,
	}, nil
}

func (r *AzureRemote) Get(ctx context.Context) (*RemoteApps, error) {
	apps := make(RemoteApps)
	pager := r.client.NewListByResourceGroupPager(r.resourceGroup, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, app := range nextResult.Value {
			managed := false
			tag, ok := app.Tags["aca.xenit.io"]
			if ok {
				if *tag == "true" {
					managed = true
				}
			}

			apps[*app.Name] = RemoteApp{
				app,
				managed,
			}
		}
	}

	return &apps, nil
}

func (r *AzureRemote) Create(ctx context.Context, name string, app armappcontainers.ContainerApp) error {
	res, err := r.client.BeginCreateOrUpdate(ctx, r.resourceGroup, name, app, &armappcontainers.ContainerAppsClientBeginCreateOrUpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create: %w", err)
	}

	_, err = res.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
		Frequency: 5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to create: %w", err)
	}

	return nil
}

func (r *AzureRemote) Update(ctx context.Context, name string, app armappcontainers.ContainerApp) error {
	return r.Create(ctx, name, app)
}

func (r *AzureRemote) Delete(ctx context.Context, name string) error {
	res, err := r.client.BeginDelete(ctx, r.resourceGroup, name, &armappcontainers.ContainerAppsClientBeginDeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	_, err = res.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
		Frequency: 5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	return nil
}

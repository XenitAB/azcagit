package remote

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
	"github.com/xenitab/aca-gitops-engine/src/config"
)

type AzureRemote struct {
	resourceGroup string
	client        *armappcontainers.ContainerAppsClient
}

var _ Remote = (*AzureRemote)(nil)

func NewAzureRemote(cfg config.Config) (*AzureRemote, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	client, err := armappcontainers.NewContainerAppsClient(cfg.SubscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	// FIXME: We need a good way to resolve if we have a working credential or not.
	//        This is just a bad workaround until something better comes around.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{"https://management.core.windows.net"}})
	if err != nil {
		return nil, err
	}

	return &AzureRemote{
		resourceGroup: cfg.ResourceGroupName,
		client:        client,
	}, nil
}

func (r *AzureRemote) List(ctx context.Context) (*RemoteApps, error) {
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

func (r *AzureRemote) CreateOrUpdate(ctx context.Context, name string, app armappcontainers.ContainerApp) error {
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

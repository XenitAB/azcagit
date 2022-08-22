package main

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
)

func getContainerAppsClient(subscriptionId string) (*armappcontainers.ContainerAppsClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	client, err := armappcontainers.NewContainerAppsClient(subscriptionId, cred, nil)
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

	return client, nil
}

type liveContainerApp struct {
	app     *armappcontainers.ContainerApp
	managed bool
}

type liveContainerApps map[string]liveContainerApp

func listContainerApps(ctx context.Context, client *armappcontainers.ContainerAppsClient, resourceGroupName string) (*liveContainerApps, error) {
	apps := make(liveContainerApps)
	pager := client.NewListByResourceGroupPager(resourceGroupName, nil)
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

			apps[*app.Name] = liveContainerApp{
				app,
				managed,
			}
		}
	}

	return &apps, nil
}

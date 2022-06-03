package main

import (
	"context"

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
	return client, nil
}

func listContainerApps(ctx context.Context, client *armappcontainers.ContainerAppsClient, resourceGroupName string) (map[string]struct{}, error) {
	apps := map[string]struct{}{}
	pager := client.NewListByResourceGroupPager(resourceGroupName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, v := range nextResult.Value {
			apps[*v.Name] = struct{}{}
		}
	}

	return apps, nil
}

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
)

func main() {
	err := run(config{
		ResourceGroupName:    "rg-aca-tenant",
		SubscriptionID:       "2a6936a5-fc30-492a-ab19-ec59068b5b96",
		ManagedEnvironmentID: "/subscriptions/2a6936a5-fc30-492a-ab19-ec59068b5b96/resourceGroups/rg-aca-platform/providers/Microsoft.App/managedEnvironments/me-container-apps",
		Location:             "west europe",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "application returned an error: %v", err)
		os.Exit(1)
	}
}

type config struct {
	ResourceGroupName    string
	SubscriptionID       string
	ManagedEnvironmentID string
	Location             string
}

func run(cfg config) error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("failed to obtain a credential: %w", err)
	}
	ctx := context.Background()
	client, err := armappcontainers.NewContainerAppsClient(cfg.SubscriptionID, cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	pager := client.NewListByResourceGroupPager(cfg.ResourceGroupName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to advance page: %w", err)
		}
		for idx, v := range nextResult.Value {
			fmt.Printf("Name #%d: %s\n", idx, *v.Name)
		}
	}

	aca, err := getAzureContainerApp(cfg)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	res, err := client.BeginCreateOrUpdate(ctx, cfg.ResourceGroupName, aca.Name, *aca.ContainerApp, &armappcontainers.ContainerAppsClientBeginCreateOrUpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create or update: %w", err)
	}

	pollRes, err := res.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
		Frequency: 5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to create or update: %w", err)
	}

	b, err := pollRes.ContainerApp.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal json poll result: %w", err)
	}

	fmt.Printf("Result: %s\n", string(b))

	return nil
}

func toPtr[T any](a T) *T {
	return &a
}

func getAzureContainerApp(cfg config) (AzureContainerApp, error) {
	b, err := os.ReadFile("./test/yaml/example.yaml")
	if err != nil {
		return AzureContainerApp{}, err
	}
	aca := AzureContainerApp{}
	err = aca.Unmarshal(b, cfg)
	if err != nil {
		return AzureContainerApp{}, err
	}
	return aca, nil
}

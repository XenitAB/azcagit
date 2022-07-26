package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
)

func main() {
	err := run(config{
		ResourceGroupName:    "rg-aca-tenant",
		SubscriptionID:       "2a6936a5-fc30-492a-ab19-ec59068b5b96",
		ManagedEnvironmentID: "/subscriptions/2a6936a5-fc30-492a-ab19-ec59068b5b96/resourceGroups/rg-aca-platform/providers/Microsoft.App/managedEnvironments/me-container-apps",
		Location:             "west europe",
		ReconcileInterval:    "10s",
		YAMLPath:             "./test/yaml",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "application returned an error: %v\n", err)
		os.Exit(1)
	}
}

type config struct {
	ResourceGroupName    string
	SubscriptionID       string
	ManagedEnvironmentID string
	Location             string
	ReconcileInterval    string
	YAMLPath             string
}

func run(cfg config) error {
	client, err := getContainerAppsClient(cfg.SubscriptionID)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return reconcile(ctx, cfg, client)
}

func reconcile(ctx context.Context, cfg config, client *armappcontainers.ContainerAppsClient) error {
	fsACAs, err := getACAs(cfg)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	liveACAs, err := listContainerApps(ctx, client, cfg.ResourceGroupName)
	if err != nil {
		return err
	}

	for name, aca := range *liveACAs {
		_, ok := (*fsACAs)[name]
		if !ok {
			if !aca.managed {
				continue
			}
			fmt.Printf("starting liveACA deletion: %s\n", name)
			err := deleteACA(ctx, name, cfg, client)
			if err != nil {
				fmt.Printf("failed liveACA deletion: %s\n", name)
				return err
			}
			fmt.Printf("finished liveACA deletion: %s\n", name)
		}
	}

	for name, aca := range *fsACAs {
		liveACA, ok := (*liveACAs)[name]
		if ok {
			if !liveACA.managed {
				return fmt.Errorf("trying to update a non-managed app: %s", name)
			}

			fmt.Printf("starting fsACA update: %s\n", name)
			err := updateACA(ctx, aca, cfg, client)
			if err != nil {
				fmt.Printf("failed fsACA update: %s\n", name)
				return err
			}
			fmt.Printf("finished fsACA update: %s\n", name)
			continue
		}

		fmt.Printf("starting fsACA creation: %s\n", name)
		err := createACA(ctx, aca, cfg, client)
		if err != nil {
			fmt.Printf("failed fsACA creation: %s\n", name)
			return err
		}
		fmt.Printf("finished fsACA creation: %s\n", name)
	}

	return nil
}

func updateACA(ctx context.Context, aca AzureContainerApp, cfg config, client *armappcontainers.ContainerAppsClient) error {
	res, err := client.BeginUpdate(ctx, cfg.ResourceGroupName, aca.Name(), *aca.Specification, &armappcontainers.ContainerAppsClientBeginUpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	_, err = res.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
		Frequency: 5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	return nil
}

func createACA(ctx context.Context, aca AzureContainerApp, cfg config, client *armappcontainers.ContainerAppsClient) error {
	res, err := client.BeginCreateOrUpdate(ctx, cfg.ResourceGroupName, aca.Name(), *aca.Specification, &armappcontainers.ContainerAppsClientBeginCreateOrUpdateOptions{})
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

func deleteACA(ctx context.Context, name string, cfg config, client *armappcontainers.ContainerAppsClient) error {
	res, err := client.BeginDelete(ctx, cfg.ResourceGroupName, name, &armappcontainers.ContainerAppsClientBeginDeleteOptions{})
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

func toPtr[T any](a T) *T {
	return &a
}

func getACAs(cfg config) (*AzureContainerApps, error) {
	yamlFiles, err := listYamlFromPath(cfg.YAMLPath)
	if err != nil {
		return nil, err
	}
	return GetAzureContainerAppFromFiles(yamlFiles, cfg)
}

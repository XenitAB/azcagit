package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"os/signal"
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
		CheckoutPath:         "/tmp/foo",
		GitUrl:               "https://github.com/simongottschlag/aca-test-yaml.git",
		GitBranch:            "main",
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
	CheckoutPath         string
	GitUrl               string
	GitBranch            string
}

func run(cfg config) error {
	client, err := getContainerAppsClient(cfg.SubscriptionID)
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	tickerInterval, err := time.ParseDuration(cfg.ReconcileInterval)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(tickerInterval)

	var hash string
	db := make(containerAppDB)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("context done")
			return nil
		case <-ticker.C:
			hash, err = reconcile(ctx, cfg, client, hash, &db)
			if err != nil {
				fmt.Printf("reconcile error: %v\n", err)
			}
			ticker.Reset(tickerInterval)
		}
	}
}

type containerAppDBEntry struct {
	modified  time.Time
	localHash string
}
type containerAppDB map[string]containerAppDBEntry

func (db *containerAppDB) set(name string, live, local *armappcontainers.ContainerApp) {
	if live == nil {
		return
	}
	if live.SystemData == nil {
		return
	}

	timestamp := live.SystemData.LastModifiedAt
	if timestamp == nil {
		if live.SystemData.CreatedAt == nil {
			return
		}
		timestamp = live.SystemData.CreatedAt
	}

	b, err := local.MarshalJSON()
	if err != nil {
		return
	}
	hash := fmt.Sprintf("%x", md5.Sum(b))

	(*db)[name] = containerAppDBEntry{
		modified:  *timestamp,
		localHash: hash,
	}
}

func (db *containerAppDB) needsUpdate(name string, live, local *armappcontainers.ContainerApp) bool {
	entry, ok := (*db)[name]
	if !ok {
		return true
	}

	if live == nil {
		return true
	}
	if live.SystemData == nil {
		return true
	}

	timestamp := live.SystemData.LastModifiedAt
	if timestamp == nil {
		if live.SystemData.CreatedAt == nil {
			return true
		}
		timestamp = live.SystemData.CreatedAt
	}

	if entry.modified != *timestamp {
		return true
	}

	b, err := local.MarshalJSON()
	if err != nil {
		return true
	}

	hash := fmt.Sprintf("%x", md5.Sum(b))
	return entry.localHash != hash
}

func reconcile(ctx context.Context, cfg config, client *armappcontainers.ContainerAppsClient, lastHash string, db *containerAppDB) (string, error) {
	fsACAs, hash, err := getACAs(ctx, cfg, lastHash)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	liveACAs, err := listContainerApps(ctx, client, cfg.ResourceGroupName)
	if err != nil {
		return "", err
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
				return "", err
			}
			fmt.Printf("finished liveACA deletion: %s\n", name)
		}
	}

	for name, aca := range *fsACAs {
		liveACA, ok := (*liveACAs)[name]
		if !db.needsUpdate(name, liveACA.app, aca.Specification) {
			fmt.Printf("Skipping update: %s\n", name)
			continue
		}
		if ok {
			if !liveACA.managed {
				return "", fmt.Errorf("trying to update a non-managed app: %s", name)
			}

			fmt.Printf("starting fsACA update: %s\n", name)
			err := createOrUpdateACA(ctx, aca, cfg, client)
			if err != nil {
				fmt.Printf("failed fsACA update: %s\n", name)
				return "", err
			}
			fmt.Printf("finished fsACA update: %s\n", name)
			continue
		}

		fmt.Printf("starting fsACA creation: %s\n", name)
		err := createOrUpdateACA(ctx, aca, cfg, client)
		if err != nil {
			fmt.Printf("failed fsACA creation: %s\n", name)
			return "", err
		}
		fmt.Printf("finished fsACA creation: %s\n", name)
	}

	newLiveACAs, err := listContainerApps(ctx, client, cfg.ResourceGroupName)
	if err != nil {
		return "", err
	}

	for name, aca := range *fsACAs {
		liveACA, ok := (*newLiveACAs)[name]
		if !ok {
			return "", fmt.Errorf("unable to locate %s after create or update", name)
		}
		db.set(name, liveACA.app, aca.Specification)
	}

	return hash, nil
}

func createOrUpdateACA(ctx context.Context, aca AzureContainerApp, cfg config, client *armappcontainers.ContainerAppsClient) error {
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

func getACAs(ctx context.Context, cfg config, lastHash string) (*AzureContainerApps, string, error) {
	yamlFiles, hash, err := checkout(ctx, cfg.CheckoutPath, cfg.GitUrl, cfg.GitBranch, lastHash)
	if err != nil {
		return nil, "", err
	}

	apps, err := GetAzureContainerAppFromFiles(yamlFiles, cfg)
	if err != nil {
		return nil, "", err
	}

	return apps, hash, nil
}

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
		ResourceGroupName:    "rg-container-apps-tenant",
		SubscriptionID:       "2a6936a5-fc30-492a-ab19-ec59068b5b96",
		ManagedEnvironmentID: "/subscriptions/2a6936a5-fc30-492a-ab19-ec59068b5b96/resourceGroups/rg-container-apps/providers/Microsoft.App/managedEnvironments/me-container-apps",
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

	app := armappcontainers.ContainerApp{
		Location: &cfg.Location,
		Identity: nil,
		Properties: &armappcontainers.ContainerAppProperties{
			ManagedEnvironmentID: &cfg.ManagedEnvironmentID,
			Configuration: &armappcontainers.Configuration{
				ActiveRevisionsMode: toPtr(armappcontainers.ActiveRevisionsModeSingle),
				Dapr: &armappcontainers.Dapr{
					AppID:   toPtr("foo-bar"),
					Enabled: toPtr(true),
				},
				Ingress: &armappcontainers.Ingress{
					External:      toPtr(true),
					TargetPort:    toPtr(int32(80)),
					AllowInsecure: toPtr(false),
				},
				Registries: nil,
				Secrets:    nil,
			},
			Template: &armappcontainers.Template{
				Containers: []*armappcontainers.Container{
					{
						Name:  toPtr("simple-hello-world-container"),
						Image: toPtr("mcr.microsoft.com/azuredocs/containerapps-helloworld:latest"),
						Resources: &armappcontainers.ContainerResources{
							CPU:    toPtr(float64(0.25)),
							Memory: toPtr(".5Gi"),
						},
					},
				},
				Scale: &armappcontainers.Scale{
					MaxReplicas: toPtr(int32(1)),
					MinReplicas: toPtr(int32(1)),
				},
			},
		},
		Tags: map[string]*string{
			"hidden-xca.xenit.io/gitops": toPtr("true"),
		},
	}

	res, err := client.BeginCreateOrUpdate(ctx, cfg.ResourceGroupName, "foo-bar", app, &armappcontainers.ContainerAppsClientBeginCreateOrUpdateOptions{})
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

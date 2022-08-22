package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/xenitab/aca-gitops-engine/src/cache"
	"github.com/xenitab/aca-gitops-engine/src/config"
	"github.com/xenitab/aca-gitops-engine/src/reconcile"
	"github.com/xenitab/aca-gitops-engine/src/remote"
	"github.com/xenitab/aca-gitops-engine/src/source"
)

func main() {
	err := run(config.Config{
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

func run(cfg config.Config) error {
	sourceClient, err := source.NewGitSource(cfg)
	if err != nil {
		return err
	}

	remoteClient, err := remote.NewAzureRemote(cfg)
	if err != nil {
		return err
	}

	cache := cache.NewCache()

	reconciler, err := reconcile.NewReconciler(sourceClient, remoteClient, cache)
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

	for {
		select {
		case <-ctx.Done():
			fmt.Println("context done")
			return nil
		case <-ticker.C:
			err := reconciler.Run(ctx)
			if err != nil {
				fmt.Printf("reconcile error: %v\n", err)
			}
			ticker.Reset(tickerInterval)
		}
	}
}

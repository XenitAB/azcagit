package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/go-logr/logr"
	"github.com/xenitab/aca-gitops-engine/src/cache"
	"github.com/xenitab/aca-gitops-engine/src/config"
	"github.com/xenitab/aca-gitops-engine/src/logger"
	"github.com/xenitab/aca-gitops-engine/src/reconcile"
	"github.com/xenitab/aca-gitops-engine/src/remote"
	"github.com/xenitab/aca-gitops-engine/src/source"
)

func main() {
	ctx, err := logger.NewLoggerContext(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "setting up logger returned an error: %v\n", err)
		os.Exit(1)
	}

	log := logr.FromContextOrDiscard(ctx)

	cfg := config.Config{
		ResourceGroupName:    "rg-aca-tenant",
		SubscriptionID:       "2a6936a5-fc30-492a-ab19-ec59068b5b96",
		ManagedEnvironmentID: "/subscriptions/2a6936a5-fc30-492a-ab19-ec59068b5b96/resourceGroups/rg-aca-platform/providers/Microsoft.App/managedEnvironments/me-container-apps",
		Location:             "west europe",
		ReconcileInterval:    "10s",
		CheckoutPath:         "/tmp/foo",
		GitUrl:               "https://github.com/simongottschlag/aca-test-yaml.git",
		GitBranch:            "main",
	}

	err = run(ctx, cfg)
	if err != nil {
		log.Error(err, "application returned an error")
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config.Config) error {
	log := logr.FromContextOrDiscard(ctx)

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

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	tickerInterval, err := time.ParseDuration(cfg.ReconcileInterval)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(tickerInterval)

	for {
		select {
		case <-ctx.Done():
			log.Info("context done, shutting down")
			return nil
		case <-ticker.C:
			err := reconciler.Run(ctx)
			if err != nil {
				log.Error(err, "reconcile error")
			}
			ticker.Reset(tickerInterval)
		}
	}
}

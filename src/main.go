package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/go-logr/logr"
	"github.com/xenitab/azcagit/src/cache"
	"github.com/xenitab/azcagit/src/config"
	"github.com/xenitab/azcagit/src/logger"
	"github.com/xenitab/azcagit/src/reconcile"
	"github.com/xenitab/azcagit/src/remote"
	"github.com/xenitab/azcagit/src/source"
)

func main() {
	ctx, err := logger.NewLoggerContext(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "setting up logger returned an error: %v\n", err)
		os.Exit(1)
	}

	log := logr.FromContextOrDiscard(ctx)
	cfg, err := config.NewConfig(os.Args[1:])
	if err != nil {
		log.Error(err, "unable to load config")
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

	_, err = sourceClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("unable to get source: %w", err)
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

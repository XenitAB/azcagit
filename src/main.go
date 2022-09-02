package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/go-logr/logr"
	"github.com/xenitab/azcagit/src/azure"
	"github.com/xenitab/azcagit/src/cache"
	"github.com/xenitab/azcagit/src/config"
	"github.com/xenitab/azcagit/src/logger"
	"github.com/xenitab/azcagit/src/notification"
	"github.com/xenitab/azcagit/src/reconcile"
	"github.com/xenitab/azcagit/src/remote"
	"github.com/xenitab/azcagit/src/secret"
	"github.com/xenitab/azcagit/src/source"
	"github.com/xenitab/azcagit/src/trigger"
	"golang.org/x/sync/errgroup"
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
		os.Exit(1)
	}

	log.Info("configuration loaded", "config", cfg.Redacted())

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

	_, _, err = sourceClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("unable to get source: %w", err)
	}

	cred, err := azure.NewAzureCredential()
	if err != nil {
		return err
	}

	remoteClient, err := remote.NewAzureRemote(cfg, cred)
	if err != nil {
		return err
	}

	secretClient, err := secret.NewKeyVaultSecret(cfg, cred)
	if err != nil {
		return err
	}

	notificationClient, err := notification.NewNotificationClient(cfg.GitUrl)
	if err != nil {
		return err
	}

	appCache := cache.NewAppCache()
	secretCache := cache.NewSecretCache()

	reconciler, err := reconcile.NewReconciler(cfg, sourceClient, remoteClient, secretClient, notificationClient, appCache, secretCache)
	if err != nil {
		return err
	}

	trig, err := trigger.NewDaprSubTrigger(cfg)
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return trig.Start()
	})

	tickerInterval, err := time.ParseDuration(cfg.ReconcileInterval)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(1 * time.Second)

	reconcile := func(triggeredBy trigger.TriggeredBy) {
		log.Info("reconcile triggered", "triggeredBy", triggeredBy)
		err := reconciler.Run(ctx)
		if err != nil {
			log.Error(err, "reconcile error")
		}
		ticker.Reset(tickerInterval)
	}

OUTER:
	for {
		select {
		case <-ctx.Done():
			log.Info("context done, shutting down")
			break OUTER
		case <-ticker.C:
			reconcile(trigger.TriggeredByTicker)
		case triggeredBy := <-trig.WaitForTrigger():
			reconcile(triggeredBy)
		}
	}

	g.Go(func() error {
		return trig.Stop()
	})

	return g.Wait()
}

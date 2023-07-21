package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/go-logr/logr"
	"github.com/xenitab/azcagit/src/azure"
	"github.com/xenitab/azcagit/src/cache"
	"github.com/xenitab/azcagit/src/config"
	"github.com/xenitab/azcagit/src/logger"
	"github.com/xenitab/azcagit/src/metrics"
	"github.com/xenitab/azcagit/src/notification"
	"github.com/xenitab/azcagit/src/reconcile"
	"github.com/xenitab/azcagit/src/remote"
	"github.com/xenitab/azcagit/src/secret"
	"github.com/xenitab/azcagit/src/source"
)

func main() {
	ctx, err := logger.NewLoggerContext(context.Background(), isDebugEnabled(os.Args))
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

	err = run(ctx, cfg)
	if err != nil {
		log.Error(err, "application returned an error")
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config.Config) error {
	log := logr.FromContextOrDiscard(ctx)

	switch {
	case cfg.ReconcileCfg != nil:
		log.Info("reconcile configuration loaded", "config", cfg.ReconcileCfg.Redacted())
		return runReconcile(ctx, *cfg.ReconcileCfg)
	case cfg.TriggerCfg != nil:
		return runTrigger(ctx, *cfg.TriggerCfg)
	}

	return fmt.Errorf("no subcommand executed")
}

func runReconcile(ctx context.Context, cfg config.ReconcileConfig) error {
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

	remoteAppClient, err := remote.NewAzureApp(cfg, cred)
	if err != nil {
		return err
	}

	remoteJobClient, err := remote.NewAzureJob(cfg, cred)
	if err != nil {
		return err
	}

	secretClient, err := secret.NewKeyVaultSecret(cfg, cred)
	if err != nil {
		return err
	}

	notificationClient, err := notification.NewNotificationClient(cfg)
	if err != nil {
		return err
	}

	metricsClient := metrics.NewAzureMetrics(cfg, cred)

	appCache := cache.NewAppCache()
	jobCache := cache.NewJobCache()
	secretCache := cache.NewSecretCache()

	reconciler, err := reconcile.NewReconciler(cfg, sourceClient, remoteAppClient, remoteJobClient, secretClient, notificationClient, metricsClient, appCache, jobCache, secretCache)
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	err = reconciler.Run(ctx)
	if err != nil {
		return fmt.Errorf("reconcile error: %w", err)
	}

	return nil
}

func runTrigger(ctx context.Context, cfg config.TriggerConfig) error {
	return nil
}

func isDebugEnabled(args []string) bool {
	for _, v := range args {
		if v == "--debug" {
			return true
		}
	}
	return false
}

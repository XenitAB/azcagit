package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
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
	cred, err := azure.NewAzureCredential()
	if err != nil {
		return err
	}

	cosmosDBClient, err := azure.NewCosmosDBClient(cfg.CosmosDBAccount, cfg.CosmosDBSqlDb, cfg.CosmosDBCacheContainer, cred)
	if err != nil {
		return err
	}

	revisionCache, err := cache.NewCosmosDBRevisionCache(cfg, cosmosDBClient)
	if err != nil {
		return err
	}

	sourceClient, err := source.NewGitSource(cfg, revisionCache)
	if err != nil {
		return err
	}

	_, _, err = sourceClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("unable to get source: %w", err)
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

	appCache, err := cache.NewCosmosDBAppCache(cfg, cosmosDBClient)
	if err != nil {
		return err
	}

	jobCache, err := cache.NewCosmosDBJobCache(cfg, cosmosDBClient)
	if err != nil {
		return err
	}

	secretCache := cache.NewInMemSecretCache()

	notificationCache, err := cache.NewCosmosDBNotificationCache(cfg, cosmosDBClient)
	if err != nil {
		return err
	}

	reconciler, err := reconcile.NewReconciler(cfg, sourceClient, remoteAppClient, remoteJobClient, secretClient, notificationClient, metricsClient, appCache, jobCache, secretCache, notificationCache)
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
	log := logr.FromContextOrDiscard(ctx)

	cred, err := azure.NewAzureCredential()
	if err != nil {
		return err
	}

	namespaceFqdn := fmt.Sprintf("%s.servicebus.windows.net", cfg.ServiceBusNamespace)
	sbClient, err := azservicebus.NewClient(namespaceFqdn, cred, nil)
	if err != nil {
		return err
	}

	defer func() {
		err := sbClient.Close(ctx)
		if err != nil {
			log.Error(err, "failed to close the service bus client")
		}
	}()

	receiver, err := sbClient.NewReceiverForQueue(cfg.ServiceBusQueue, &azservicebus.ReceiverOptions{})
	if err != nil {
		return err
	}

	peekedMessages, err := receiver.PeekMessages(ctx, 100, &azservicebus.PeekMessagesOptions{})
	if err != nil {
		return err
	}

	if len(peekedMessages) > 0 {
		receivedMessages, err := receiver.ReceiveMessages(ctx, 100, &azservicebus.ReceiveMessagesOptions{})
		if err != nil {
			return err
		}
		for _, msg := range receivedMessages {
			err := receiver.CompleteMessage(ctx, msg, &azservicebus.CompleteMessageOptions{})
			if err != nil {
				return err
			}
		}
	}

	jobClient, err := armappcontainers.NewJobsClient(cfg.SubscriptionID, cred, nil)
	if err != nil {
		return err
	}

	res, err := jobClient.BeginStart(ctx, cfg.ResourceGroupName, cfg.JobName, &armappcontainers.JobsClientBeginStartOptions{})
	if err != nil {
		return err
	}

	_, err = res.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
		Frequency: 5 * time.Second,
	})

	if err != nil {
		return err
	}

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

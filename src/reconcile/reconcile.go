package reconcile

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"github.com/xenitab/azcagit/src/cache"
	"github.com/xenitab/azcagit/src/config"
	"github.com/xenitab/azcagit/src/notification"
	"github.com/xenitab/azcagit/src/remote"
	"github.com/xenitab/azcagit/src/secret"
	"github.com/xenitab/azcagit/src/source"
)

type Reconciler struct {
	cfg                       config.Config
	sourceClient              source.Source
	remoteClient              remote.Remote
	secretClient              secret.Secret
	notificationClient        notification.Notification
	appCache                  *cache.AppCache
	secretCache               *cache.SecretCache
	previousNotificationEvent notification.NotificationEvent
}

func NewReconciler(cfg config.Config, sourceClient source.Source, remoteClient remote.Remote, secretClient secret.Secret, notificationClient notification.Notification, appCache *cache.AppCache, secretCache *cache.SecretCache) (*Reconciler, error) {
	previousNotificationEvent := notification.NotificationEvent{}
	return &Reconciler{
		cfg,
		sourceClient,
		remoteClient,
		secretClient,
		notificationClient,
		appCache,
		secretCache,
		previousNotificationEvent,
	}, nil
}

func (r *Reconciler) Run(ctx context.Context) error {
	var result *multierror.Error

	revision, reconcileErr := r.run(ctx)
	if reconcileErr != nil {
		result = multierror.Append(reconcileErr, result)
	}

	err := r.sendNotification(ctx, revision, reconcileErr)
	if err != nil {
		result = multierror.Append(err, result)
	}

	return result.ErrorOrNil()
}

func (r *Reconciler) run(ctx context.Context) (string, error) {
	sourceApps, revision, err := r.getSourceApps(ctx)
	if err != nil {
		return revision, err
	}

	err = r.populateSourceAppsSecrets(ctx, sourceApps)
	if err != nil {
		return revision, err
	}

	err = r.populateSourceAppsRegistries(sourceApps)
	if err != nil {
		return revision, err
	}

	remoteApps, err := r.getRemoteApps(ctx)
	if err != nil {
		return revision, err
	}

	err = r.deleteAppsIfNeeded(ctx, sourceApps, remoteApps)
	if err != nil {
		return revision, err
	}

	err = r.createOrUpdateAppsIfNeeded(ctx, sourceApps, remoteApps)
	if err != nil {
		return revision, err
	}

	err = r.updateCache(ctx, sourceApps)
	if err != nil {
		return revision, err
	}

	return revision, nil
}

func (r *Reconciler) getSourceApps(ctx context.Context) (*source.SourceApps, string, error) {
	sourceApps, revision, err := r.sourceClient.Get(ctx)
	if err != nil {
		return nil, revision, fmt.Errorf("failed to get sourceApps: %w", err)
	}

	if sourceApps == nil {
		return nil, revision, fmt.Errorf("sourceApps is nil")
	}

	if sourceApps.Error() != nil {
		return nil, revision, fmt.Errorf("sourceApps contains errors, stopping reconciliation: %w", sourceApps.Error())
	}

	return sourceApps, revision, nil
}

func (r *Reconciler) getRemoteApps(ctx context.Context) (*remote.RemoteApps, error) {
	remoteApps, err := r.remoteClient.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get remoteApps: %w", err)
	}

	if remoteApps == nil {
		return nil, fmt.Errorf("remoteApps is nil")
	}

	return remoteApps, nil
}

func (r *Reconciler) deleteAppsIfNeeded(ctx context.Context, sourceApps *source.SourceApps, remoteApps *remote.RemoteApps) error {
	log := logr.FromContextOrDiscard(ctx)

	for _, name := range remoteApps.GetSortedNames() {
		if sourceApps.Error() != nil {
			log.Error(fmt.Errorf("delete disabled"), "no remoteApps will be deleted while sourceApps contains errors")
			break
		}
		remoteApp, _ := remoteApps.Get(name)
		_, ok := sourceApps.Get(name)
		if !ok {
			if !remoteApp.Managed {
				continue
			}
			err := r.remoteClient.Delete(ctx, name)
			if err != nil {
				return err
			}
			log.Info("deleted remoteApp", "app", name)
		}
	}

	return nil
}

func (r *Reconciler) createOrUpdateAppsIfNeeded(ctx context.Context, sourceApps *source.SourceApps, remoteApps *remote.RemoteApps) error {
	log := logr.FromContextOrDiscard(ctx)

	for _, name := range sourceApps.GetSortedNames() {
		sourceApp, _ := sourceApps.Get(name)
		remoteApp, ok := remoteApps.Get(name)
		needsUpdate, updateReason := r.appCache.NeedsUpdate(name, remoteApp.App, sourceApp.Specification.App)
		if !needsUpdate {
			log.Info("skipping update, no changes", "app", name)
			continue
		}
		if ok {
			if !remoteApp.Managed {
				return fmt.Errorf("trying to update a non-managed app: %s", name)
			}

			err := r.remoteClient.Update(ctx, name, *sourceApp.Specification.App)
			if err != nil {
				return fmt.Errorf("failed to update %s: %w", name, err)
			}
			log.Info("updated remoteApp", "app", name, "reason", updateReason)
			continue
		}

		err := r.remoteClient.Create(ctx, name, *sourceApp.Specification.App)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", name, err)
		}
		log.Info("created remoteApp", "app", name, "reason", updateReason)
	}

	return nil
}

func (r *Reconciler) updateCache(ctx context.Context, sourceApps *source.SourceApps) error {
	newRemoteApps, err := r.remoteClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get new remoteApps: %w", err)
	}

	for _, name := range sourceApps.GetSortedNames() {
		sourceApp, _ := sourceApps.Get(name)
		remoteApp, ok := newRemoteApps.Get(name)
		if !ok {
			return fmt.Errorf("unable to locate %s after create or update", name)
		}
		r.appCache.Set(name, remoteApp.App, sourceApp.Specification.App)
	}

	return nil
}

func (r *Reconciler) populateSourceAppsSecrets(ctx context.Context, sourceApps *source.SourceApps) error {
	secretItems, err := r.secretClient.ListItems(ctx)
	if err != nil {
		return err
	}

	for _, secretName := range sourceApps.GetUniqueRemoteSecretNames() {
		item, ok := secretItems.Get(secretName)
		if !ok {
			return fmt.Errorf("secret not found %q", secretName)
		}

		if r.secretCache.NeedsUpdate(secretName, item.LastChange()) {
			secretValue, changedAt, err := r.secretClient.Get(ctx, secretName)
			if err != nil {
				return err
			}
			r.secretCache.Set(secretName, secretValue, changedAt)
		}
	}

	for _, name := range sourceApps.GetSortedNames() {
		app, _ := sourceApps.Get(name)
		for i, remoteSecret := range app.GetRemoteSecrets() {
			if !remoteSecret.Valid() {
				return fmt.Errorf("secret %d for app %q not valid", i, name)
			}

			secretValue, ok := r.secretCache.Get(*remoteSecret.RemoteSecretName)
			if !ok {
				return fmt.Errorf("unable to get secret %d for app %q from cache", i, name)
			}

			err = sourceApps.SetAppSecret(name, *remoteSecret.AppSecretName, secretValue)
			if err != nil {
				return fmt.Errorf("unable to set secret %q for app %q", *remoteSecret.AppSecretName, name)
			}
		}
	}

	return nil
}

func (r *Reconciler) populateSourceAppsRegistries(sourceApps *source.SourceApps) error {
	if r.cfg.ContainerRegistryServer == "" && r.cfg.ContainerRegistryUsername == "" && r.cfg.ContainerRegistryPassword == "" {
		return nil
	}

	if r.cfg.ContainerRegistryServer != "" && (r.cfg.ContainerRegistryUsername == "" || r.cfg.ContainerRegistryPassword == "") {
		return fmt.Errorf("all of container registry server, username and password needs to be set")
	}

	for _, name := range sourceApps.GetSortedNames() {
		err := sourceApps.SetAppRegistry(name, r.cfg.ContainerRegistryServer, r.cfg.ContainerRegistryUsername, r.cfg.ContainerRegistryPassword)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Reconciler) sendNotification(ctx context.Context, revision string, reconcileErr error) error {
	log := logr.FromContextOrDiscard(ctx)
	log.V(1).Info("sendNotification invoked", "revision", revision, "reconcileErr", reconcileErr)

	if revision == "" {
		log.V(1).Error(fmt.Errorf("revision empty"), "unable to send notification when revision is empty")
		return fmt.Errorf("unable to send notifications when revision is empty")
	}

	description := "reconcile succeeded"
	state := notification.NotificationStateSuccess
	if reconcileErr != nil {
		description = reconcileErr.Error()
		state = notification.NotificationStateFailure
	}

	location := strings.ReplaceAll(r.cfg.Location, " ", "")
	resourceGroup := r.cfg.ResourceGroupName
	name := strings.ToLower(fmt.Sprintf("%s/%s", location, resourceGroup))
	event := notification.NotificationEvent{
		Revision:    revision,
		State:       state,
		Name:        name,
		Description: description,
	}

	if r.previousNotificationEvent.Equal(event) {
		log.V(1).Info("skipping notification, events are equal", "current_event", event, "previous_event", r.previousNotificationEvent)
		return nil
	}

	r.previousNotificationEvent = event

	err := r.notificationClient.Send(ctx, event)
	if err != nil {
		log.V(1).Error(err, "unable to send event, received error", "event", event)
		return err
	}

	log.V(1).Info("event sent", "event", event)

	return nil
}

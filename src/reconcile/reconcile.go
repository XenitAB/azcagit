package reconcile

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/xenitab/azcagit/src/cache"
	"github.com/xenitab/azcagit/src/remote"
	"github.com/xenitab/azcagit/src/source"
)

type Reconciler struct {
	sourceClient source.Source
	remoteClient remote.Remote
	cache        *cache.Cache
}

func NewReconciler(sourceClient source.Source, remoteClient remote.Remote, cache *cache.Cache) (*Reconciler, error) {
	return &Reconciler{
		sourceClient,
		remoteClient,
		cache,
	}, nil
}

func (r *Reconciler) Run(ctx context.Context) error {
	sourceApps, err := r.getSourceApps(ctx)
	if err != nil {
		return err
	}

	remoteApps, err := r.getRemoteApps(ctx)
	if err != nil {
		return err
	}

	err = r.deleteAppsIfNeeded(ctx, sourceApps, remoteApps)
	if err != nil {
		return err
	}

	err = r.createOrUpdateAppsIfNeeded(ctx, sourceApps, remoteApps)
	if err != nil {
		return err
	}

	err = r.updateCache(ctx, sourceApps)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) getSourceApps(ctx context.Context) (*source.SourceApps, error) {
	sourceApps, err := r.sourceClient.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sourceApps: %w", err)
	}

	if sourceApps == nil {
		return nil, fmt.Errorf("sourceApps is nil")
	}

	if sourceApps.Error() != nil {
		return nil, fmt.Errorf("sourceApps contains errors, stopping reconciliation: %w", sourceApps.Error())
	}

	return sourceApps, nil
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
		needsUpdate, updateReason := r.cache.NeedsUpdate(name, remoteApp.App, sourceApp.Specification)
		if !needsUpdate {
			log.Info("skipping update, no changes", "app", name)
			continue
		}
		if ok {
			if !remoteApp.Managed {
				return fmt.Errorf("trying to update a non-managed app: %s", name)
			}

			err := r.remoteClient.Update(ctx, name, *sourceApp.Specification)
			if err != nil {
				return fmt.Errorf("failed to update %s: %w", name, err)
			}
			log.Info("updated remoteApp", "app", name, "reason", updateReason)
			continue
		}

		err := r.remoteClient.Create(ctx, name, *sourceApp.Specification)
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
		r.cache.Set(name, remoteApp.App, sourceApp.Specification)
	}

	return nil
}

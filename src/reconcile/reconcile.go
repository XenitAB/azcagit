package reconcile

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/xenitab/aca-gitops-engine/src/cache"
	"github.com/xenitab/aca-gitops-engine/src/remote"
	"github.com/xenitab/aca-gitops-engine/src/source"
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
	log := logr.FromContextOrDiscard(ctx)

	sourceApps, err := r.sourceClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get sourceApps: %w", err)
	}

	if sourceApps == nil {
		return fmt.Errorf("sourceApps is nil")
	}

	if sourceApps.Error() != nil {
		log.Error(sourceApps.Error(), "sourceApps contains errors")
	}

	remoteApps, err := r.remoteClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get remoteApps: %w", err)
	}

	if remoteApps == nil {
		return fmt.Errorf("remoteApps is nil")
	}

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

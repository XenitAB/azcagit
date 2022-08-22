package reconcile

import (
	"context"
	"fmt"

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
	sourceApps, err := r.sourceClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	remoteApps, err := r.remoteClient.List(ctx)
	if err != nil {
		return err
	}

	for name, rmtApp := range *remoteApps {
		_, ok := (*sourceApps)[name]
		if !ok {
			if !rmtApp.Managed() {
				continue
			}
			fmt.Printf("starting remote app deletion: %s\n", name)
			err := r.remoteClient.Delete(ctx, name)
			if err != nil {
				fmt.Printf("failed remote app deletion: %s\n", name)
				return err
			}
			fmt.Printf("finished remote app deletion: %s\n", name)
		}
	}

	for name, sourceApp := range *sourceApps {
		remoteApp, ok := (*remoteApps)[name]
		if !r.cache.NeedsUpdate(name, remoteApp.App(), sourceApp.Specification) {
			fmt.Printf("Skipping update: %s\n", name)
			continue
		}
		if ok {
			if !remoteApp.Managed() {
				return fmt.Errorf("trying to update a non-managed app: %s", name)
			}

			fmt.Printf("starting update: %s\n", name)
			err := r.remoteClient.CreateOrUpdate(ctx, name, *sourceApp.Specification)
			if err != nil {
				fmt.Printf("failed update: %s\n", name)
				return err
			}
			fmt.Printf("finished update: %s\n", name)
			continue
		}

		fmt.Printf("starting creation: %s\n", name)
		err := r.remoteClient.CreateOrUpdate(ctx, name, *sourceApp.Specification)
		if err != nil {
			fmt.Printf("failed creation: %s\n", name)
			return err
		}
		fmt.Printf("finished creation: %s\n", name)
	}

	newRemoteApps, err := r.remoteClient.List(ctx)
	if err != nil {
		return err
	}

	for name, sourceApp := range *sourceApps {
		remoteApp, ok := (*newRemoteApps)[name]
		if !ok {
			return fmt.Errorf("unable to locate %s after create or update", name)
		}
		r.cache.Set(name, remoteApp.App(), sourceApp.Specification)
	}

	return nil
}

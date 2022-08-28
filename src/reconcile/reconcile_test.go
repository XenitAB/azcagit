package reconcile

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
	"github.com/stretchr/testify/require"
	"github.com/xenitab/azcagit/src/cache"
	"github.com/xenitab/azcagit/src/config"
	"github.com/xenitab/azcagit/src/remote"
	"github.com/xenitab/azcagit/src/source"
)

func TestReconciler(t *testing.T) {
	sourceClient, err := source.NewInMemSource(config.Config{})
	require.NoError(t, err)
	remoteClient, err := remote.NewInMemRemote(config.Config{})
	require.NoError(t, err)
	cache := cache.NewCache()

	ctx := context.Background()

	reconciler, err := NewReconciler(sourceClient, remoteClient, cache)
	require.NoError(t, err)

	resetClients := func() {
		sourceClient.GetResponse(nil, nil)
		remoteClient.GetFirstResponse(nil, nil)
		remoteClient.GetSecondResponse(nil, nil)
		remoteClient.ResetGetSecond()
		remoteClient.CreateResponse(nil)
		remoteClient.UpdateResponse(nil)
		remoteClient.DeleteResponse(nil)
		remoteClient.ResetActions()
	}

	// everything is nil
	func() {
		defer resetClients()
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "sourceApps is nil")
	}()

	// sourceClient.Get() returns error
	func() {
		defer resetClients()
		sourceClient.GetResponse(nil, fmt.Errorf("foobar"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "failed to get sourceApps: foobar")
	}()

	// sourceClient.Get() returns empty SourceApps without error
	// first remoteClient.Get() returns nil
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.SourceApps{}, nil)
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "remoteApps is nil")
	}()

	// sourceClient.Get() returns empty SourceApps without error
	// first remoteClient.Get() returns error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.SourceApps{}, nil)
		remoteClient.GetFirstResponse(nil, fmt.Errorf("foobar"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "failed to get remoteApps: foobar")
	}()

	// sourceClient.Get() returns empty SourceApps without error
	// first remoteClient.Get() returns empty RemoteApps without error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.SourceApps{}, nil)
		remoteClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		err := reconciler.Run(ctx)
		require.NoError(t, err)
	}()

	// sourceClient.Get() returns one SourceApp without error
	// first remoteClient.Get() returns empty RemoteApps without error
	// second remoteClient.Get() returns nil RemoteApps with error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.SourceApps{
			"foo": source.SourceApp{
				Kind:       "AzureContainerApp",
				APIVersion: "aca.xenit.io/v1alpha1",
				Metadata: map[string]string{
					"name": "foo",
				},
				Specification: &armappcontainers.ContainerApp{},
			},
		}, nil)
		remoteClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		remoteClient.GetSecondResponse(nil, fmt.Errorf("foobar second"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "failed to get new remoteApps: foobar second")
		actions := remoteClient.Actions()
		require.Len(t, actions, 1)
		require.Equal(t, actions[0].Name, "foo")
		require.Equal(t, actions[0].Action, remote.InMemRemoteActionsCreate)
	}()

	// sourceClient.Get() returns one SourceApp without error
	// first remoteClient.Get() returns empty RemoteApps without error
	// second remoteClient.Get() returns one RemoteApp without error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.SourceApps{
			"foo": source.SourceApp{
				Kind:       "AzureContainerApp",
				APIVersion: "aca.xenit.io/v1alpha1",
				Metadata: map[string]string{
					"name": "foo",
				},
				Specification: &armappcontainers.ContainerApp{},
			},
		}, nil)
		remoteClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		remoteClient.GetSecondResponse(&remote.RemoteApps{
			"foo": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.NoError(t, err)
		actions := remoteClient.Actions()
		require.Len(t, actions, 1)
		require.Equal(t, actions[0].Name, "foo")
		require.Equal(t, actions[0].Action, remote.InMemRemoteActionsCreate)
	}()

	// sourceClient.Get() returns two SourceApps without error
	// first remoteClient.Get() returns empty RemoteApps without error
	// second remoteClient.Get() returns one RemoteApp without error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.SourceApps{
			"foo1": source.SourceApp{
				Kind:       "AzureContainerApp",
				APIVersion: "aca.xenit.io/v1alpha1",
				Metadata: map[string]string{
					"name": "foo1",
				},
				Specification: &armappcontainers.ContainerApp{},
			},
			"foo2": source.SourceApp{
				Kind:       "AzureContainerApp",
				APIVersion: "aca.xenit.io/v1alpha1",
				Metadata: map[string]string{
					"name": "foo2",
				},
				Specification: &armappcontainers.ContainerApp{},
			},
		}, nil)
		remoteClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		remoteClient.GetSecondResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "unable to locate foo2 after create or update")
		actions := remoteClient.Actions()
		require.Len(t, actions, 2)
		require.Equal(t, actions[0].Name, "foo1")
		require.Equal(t, actions[0].Action, remote.InMemRemoteActionsCreate)
		require.Equal(t, actions[1].Name, "foo2")
		require.Equal(t, actions[1].Action, remote.InMemRemoteActionsCreate)
	}()

	// sourceClient.Get() returns two SourceApps without error
	// first remoteClient.Get() returns empty RemoteApps without error
	// second remoteClient.Get() returns two RemoteApps without error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.SourceApps{
			"foo1": source.SourceApp{
				Kind:       "AzureContainerApp",
				APIVersion: "aca.xenit.io/v1alpha1",
				Metadata: map[string]string{
					"name": "foo1",
				},
				Specification: &armappcontainers.ContainerApp{},
			},
			"foo2": source.SourceApp{
				Kind:       "AzureContainerApp",
				APIVersion: "aca.xenit.io/v1alpha1",
				Metadata: map[string]string{
					"name": "foo2",
				},
				Specification: &armappcontainers.ContainerApp{},
			},
		}, nil)
		remoteClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		remoteClient.GetSecondResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
			"foo2": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.NoError(t, err)
		actions := remoteClient.Actions()
		require.Len(t, actions, 2)
		require.Equal(t, actions[0].Name, "foo1")
		require.Equal(t, actions[0].Action, remote.InMemRemoteActionsCreate)
		require.Equal(t, actions[1].Name, "foo2")
		require.Equal(t, actions[1].Action, remote.InMemRemoteActionsCreate)
	}()

	// sourceClient.Get() returns one SourceApps without error
	// first remoteClient.Get() returns two RemoteApps without error
	// second remoteClient.Get() returns one RemoteApps without error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.SourceApps{
			"foo1": source.SourceApp{
				Kind:       "AzureContainerApp",
				APIVersion: "aca.xenit.io/v1alpha1",
				Metadata: map[string]string{
					"name": "foo1",
				},
				Specification: &armappcontainers.ContainerApp{},
			},
		}, nil)
		remoteClient.GetFirstResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
			"foo2": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		remoteClient.GetSecondResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.NoError(t, err)
		actions := remoteClient.Actions()
		require.Len(t, actions, 2)
		require.Equal(t, actions[0].Name, "foo2")
		require.Equal(t, actions[0].Action, remote.InMemRemoteActionsDelete)
		require.Equal(t, actions[1].Name, "foo1")
		require.Equal(t, actions[1].Action, remote.InMemRemoteActionsUpdate)
	}()

	// verify that if any sourceApp contains parsing error, reconciliation stops
	// sourceClient.Get() returns one SourceApp with error
	// first remoteClient.Get() returns two RemoteApps without error
	// second remoteClient.Get() returns one RemoteApps without error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.SourceApps{
			"foo1": source.SourceApp{
				Kind:       "AzureContainerApp",
				APIVersion: "aca.xenit.io/v1alpha1",
				Metadata: map[string]string{
					"name": "foo1",
				},
				Specification: &armappcontainers.ContainerApp{},
				Err:           fmt.Errorf("foobar"),
			},
		}, nil)
		remoteClient.GetFirstResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
			"foo2": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		remoteClient.GetSecondResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "sourceApps contains errors, stopping reconciliation")
		actions := remoteClient.Actions()
		require.Len(t, actions, 0)
	}()

	// test cache
	// sourceClient.Get() returns one SourceApp without error
	// first remoteClient.Get() returns one RemoteApp without error
	// second remoteClient.Get() returns one RemoteApp without error
	func() {
		defer resetClients()
		now := time.Now()
		later := time.Now().Add(1 * time.Minute)
		sourceApp1 := source.SourceApp{
			Kind:       "AzureContainerApp",
			APIVersion: "aca.xenit.io/v1alpha1",
			Metadata: map[string]string{
				"name": "foo1",
			},
			Specification: &armappcontainers.ContainerApp{
				Name: toPtr("foo1"),
			},
		}
		sourceApp1Updated := source.SourceApp{
			Kind:       "AzureContainerApp",
			APIVersion: "aca.xenit.io/v1alpha1",
			Metadata: map[string]string{
				"name": "foo1",
			},
			Specification: &armappcontainers.ContainerApp{
				Name: toPtr("foo1"),
				Tags: map[string]*string{
					"foo": toPtr("bar"),
				},
			},
		}
		remoteApp1 := remote.RemoteApp{
			App: &armappcontainers.ContainerApp{
				SystemData: &armappcontainers.SystemData{
					LastModifiedAt: &now,
				},
			},
			Managed: true,
		}
		remoteApp1Later := remote.RemoteApp{
			App: &armappcontainers.ContainerApp{
				SystemData: &armappcontainers.SystemData{
					LastModifiedAt: &later,
				},
			},
			Managed: true,
		}
		sourceClient.GetResponse(&source.SourceApps{
			"foo1": sourceApp1,
		}, nil)
		remoteClient.GetFirstResponse(&remote.RemoteApps{
			"foo1": remoteApp1,
		}, nil)
		remoteClient.GetSecondResponse(&remote.RemoteApps{
			"foo1": remoteApp1,
		}, nil)

		// run once and cache
		{
			err := reconciler.Run(ctx)
			require.NoError(t, err)
			actions := remoteClient.Actions()
			require.Len(t, actions, 1)
			require.Equal(t, actions[0].Name, "foo1")
			require.Equal(t, actions[0].Action, remote.InMemRemoteActionsUpdate)
			remoteClient.ResetActions()
		}

		// verify no actions taken
		{
			err := reconciler.Run(ctx)
			require.NoError(t, err)
			actions := remoteClient.Actions()
			require.Len(t, actions, 0)
		}

		// verify that update is made if cache is outdated
		{
			remoteClient.GetFirstResponse(&remote.RemoteApps{
				"foo1": remoteApp1Later,
			}, nil)
			remoteClient.GetSecondResponse(&remote.RemoteApps{
				"foo1": remoteApp1Later,
			}, nil)
			err := reconciler.Run(ctx)
			require.NoError(t, err)
			actions := remoteClient.Actions()
			require.Len(t, actions, 1)
			require.Equal(t, actions[0].Name, "foo1")
			require.Equal(t, actions[0].Action, remote.InMemRemoteActionsUpdate)
			remoteClient.ResetActions()
		}

		// verify that update is made if source app changed
		{
			sourceClient.GetResponse(&source.SourceApps{
				"foo1": sourceApp1Updated,
			}, nil)
			remoteClient.GetFirstResponse(&remote.RemoteApps{
				"foo1": remoteApp1Later,
			}, nil)
			remoteClient.GetSecondResponse(&remote.RemoteApps{
				"foo1": remoteApp1Later,
			}, nil)
			err := reconciler.Run(ctx)
			require.NoError(t, err)
			actions := remoteClient.Actions()
			require.Len(t, actions, 1)
			require.Equal(t, actions[0].Name, "foo1")
			require.Equal(t, actions[0].Action, remote.InMemRemoteActionsUpdate)
			remoteClient.ResetActions()
		}
	}()

	// do not delete unmanaged remote apps
	// sourceClient.Get() returns empty SourceApps without error
	// first remoteClient.Get() returns one RemoteApp (non managed) without error
	// second remoteClient.Get() returns one RemoteApp (non managed) without error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.SourceApps{}, nil)
		remoteClient.GetFirstResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: false,
			},
		}, nil)
		remoteClient.GetSecondResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: false,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.NoError(t, err)
		actions := remoteClient.Actions()
		require.Len(t, actions, 0)
	}()

	// delete failure
	// sourceClient.Get() returns empty SourceApps without error
	// first remoteClient.Get() returns one RemoteApp without error
	// remoteClient.Delete() returns error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.SourceApps{}, nil)
		remoteClient.GetFirstResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		remoteClient.DeleteResponse(fmt.Errorf("delete foobar"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "delete foobar")
	}()

	// update failure
	// sourceClient.Get() returns one SourceApp without error
	// first remoteClient.Get() returns one RemoteApp without error
	// remoteClient.Set() returns error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.SourceApps{
			"foo1": source.SourceApp{
				Kind:       "AzureContainerApp",
				APIVersion: "aca.xenit.io/v1alpha1",
				Metadata: map[string]string{
					"name": "foo1",
				},
				Specification: &armappcontainers.ContainerApp{},
			},
		}, nil)
		remoteClient.GetFirstResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		remoteClient.UpdateResponse(fmt.Errorf("update foobar"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "update foobar")
	}()

	// new app failure
	// sourceClient.Get() returns one SourceApp without error
	// first remoteClient.Get() returns empty RemoteApps without error
	// remoteClient.Set() returns error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.SourceApps{
			"foo1": source.SourceApp{
				Kind:       "AzureContainerApp",
				APIVersion: "aca.xenit.io/v1alpha1",
				Metadata: map[string]string{
					"name": "foo1",
				},
				Specification: &armappcontainers.ContainerApp{},
			},
		}, nil)
		remoteClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		remoteClient.CreateResponse(fmt.Errorf("new app foobar"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "new app foobar")
	}()
}

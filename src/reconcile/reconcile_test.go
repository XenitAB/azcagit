package reconcile

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
	"github.com/stretchr/testify/require"
	"github.com/xenitab/azcagit/src/cache"
	"github.com/xenitab/azcagit/src/config"
	"github.com/xenitab/azcagit/src/metrics"
	"github.com/xenitab/azcagit/src/notification"
	"github.com/xenitab/azcagit/src/remote"
	"github.com/xenitab/azcagit/src/secret"
	"github.com/xenitab/azcagit/src/source"
)

const defaultFakeRevision = "6ffa5a7b2da7dc37e186e2581a903e325bbd38be"

func TestReconciler(t *testing.T) {
	sourceClient := source.NewInMemSource()
	remoteAppClient := remote.NewInMemApp()
	remoteJobClient := remote.NewInMemJob()
	secretClient := secret.NewInMemSecret()
	notificationClient := notification.NewInMemNotification()
	metricsClient := metrics.NewInMemMetrics()
	appCache := cache.NewAppCache()
	jobCache := cache.NewJobCache()
	secretCache := cache.NewSecretCache()

	ctx := context.Background()

	reconciler, err := NewReconciler(config.Config{}, sourceClient, remoteAppClient, remoteJobClient, secretClient, notificationClient, metricsClient, appCache, jobCache, secretCache)
	require.NoError(t, err)

	resetClients := func() {
		sourceClient.GetResponse(nil, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(nil, nil)
		remoteAppClient.GetSecondResponse(nil, nil)
		remoteAppClient.ResetGetSecond()
		remoteAppClient.CreateResponse(nil)
		remoteAppClient.UpdateResponse(nil)
		remoteAppClient.DeleteResponse(nil)
		remoteAppClient.ResetActions()
		secretClient.Reset()
		notificationClient.SendResponse(nil)
		notificationClient.ResetNotifications()
		metricsClient.Reset()
		reconciler.previousNotificationEvent = notification.NotificationEvent{}
	}

	// everything is nil
	func() {
		defer resetClients()
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "sources is nil")
	}()

	// sourceClient.Get() returns error
	func() {
		defer resetClients()
		sourceClient.GetResponse(nil, defaultFakeRevision, fmt.Errorf("foobar"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "failed to get sourceApps: foobar")
	}()

	// sourceClient.Get() returns empty SourceApps without error
	// first remoteClient.Get() returns nil
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{Apps: &source.SourceApps{}}, defaultFakeRevision, nil)
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "remoteApps is nil")
	}()

	// sourceClient.Get() returns empty SourceApps without error
	// first remoteClient.Get() returns error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{Apps: &source.SourceApps{}}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(nil, fmt.Errorf("foobar"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "failed to get remoteApps: foobar")
	}()

	// sourceClient.Get() returns empty SourceApps without error
	// first remoteClient.Get() returns empty RemoteApps without error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{Apps: &source.SourceApps{}}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		err := reconciler.Run(ctx)
		require.NoError(t, err)
	}()

	// sourceClient.Get() returns one SourceApp without error
	// first remoteClient.Get() returns empty RemoteApps without error
	// second remoteClient.Get() returns nil RemoteApps with error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Apps: &source.SourceApps{
				"foo": source.SourceApp{
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &source.SourceAppSpecification{
						App: &armappcontainers.ContainerApp{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		remoteAppClient.GetSecondResponse(nil, fmt.Errorf("foobar second"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "failed to get new remoteApps: foobar second")
		actions := remoteAppClient.Actions()
		require.Len(t, actions, 1)
		require.Equal(t, actions[0].Name, "foo")
		require.Equal(t, actions[0].Action, remote.InMemAppActionsCreate)
	}()

	// sourceClient.Get() returns one SourceApp without error
	// first remoteClient.Get() returns empty RemoteApps without error
	// second remoteClient.Get() returns one RemoteApp without error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Apps: &source.SourceApps{
				"foo": source.SourceApp{
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &source.SourceAppSpecification{
						App: &armappcontainers.ContainerApp{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		remoteAppClient.GetSecondResponse(&remote.RemoteApps{
			"foo": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.NoError(t, err)
		actions := remoteAppClient.Actions()
		require.Len(t, actions, 1)
		require.Equal(t, actions[0].Name, "foo")
		require.Equal(t, actions[0].Action, remote.InMemAppActionsCreate)
	}()

	// sourceClient.Get() returns two SourceApps without error
	// first remoteClient.Get() returns empty RemoteApps without error
	// second remoteClient.Get() returns one RemoteApp without error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Apps: &source.SourceApps{
				"foo1": source.SourceApp{
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo1",
					},
					Specification: &source.SourceAppSpecification{
						App: &armappcontainers.ContainerApp{},
					},
				},
				"foo2": source.SourceApp{
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo2",
					},
					Specification: &source.SourceAppSpecification{
						App: &armappcontainers.ContainerApp{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		remoteAppClient.GetSecondResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "unable to locate foo2 after create or update")
		actions := remoteAppClient.Actions()
		require.Len(t, actions, 2)
		require.Equal(t, actions[0].Name, "foo1")
		require.Equal(t, actions[0].Action, remote.InMemAppActionsCreate)
		require.Equal(t, actions[1].Name, "foo2")
		require.Equal(t, actions[1].Action, remote.InMemAppActionsCreate)
	}()

	// sourceClient.Get() returns two SourceApps without error
	// first remoteClient.Get() returns empty RemoteApps without error
	// second remoteClient.Get() returns two RemoteApps without error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Apps: &source.SourceApps{
				"foo1": source.SourceApp{
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo1",
					},
					Specification: &source.SourceAppSpecification{
						App: &armappcontainers.ContainerApp{},
					},
				},
				"foo2": source.SourceApp{
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo2",
					},
					Specification: &source.SourceAppSpecification{
						App: &armappcontainers.ContainerApp{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		remoteAppClient.GetSecondResponse(&remote.RemoteApps{
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
		actions := remoteAppClient.Actions()
		require.Len(t, actions, 2)
		require.Equal(t, actions[0].Name, "foo1")
		require.Equal(t, actions[0].Action, remote.InMemAppActionsCreate)
		require.Equal(t, actions[1].Name, "foo2")
		require.Equal(t, actions[1].Action, remote.InMemAppActionsCreate)
	}()

	// sourceClient.Get() returns one SourceApps without error
	// first remoteClient.Get() returns two RemoteApps without error
	// second remoteClient.Get() returns one RemoteApps without error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Apps: &source.SourceApps{
				"foo1": source.SourceApp{
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo1",
					},
					Specification: &source.SourceAppSpecification{
						App: &armappcontainers.ContainerApp{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
			"foo2": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		remoteAppClient.GetSecondResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.NoError(t, err)
		actions := remoteAppClient.Actions()
		require.Len(t, actions, 2)
		require.Equal(t, actions[0].Name, "foo2")
		require.Equal(t, actions[0].Action, remote.InMemAppActionsDelete)
		require.Equal(t, actions[1].Name, "foo1")
		require.Equal(t, actions[1].Action, remote.InMemAppActionsUpdate)
	}()

	// verify that if any sourceApp contains parsing error, reconciliation stops
	// sourceClient.Get() returns one SourceApp with error
	// first remoteClient.Get() returns two RemoteApps without error
	// second remoteClient.Get() returns one RemoteApps without error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Apps: &source.SourceApps{
				"foo1": source.SourceApp{
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo1",
					},
					Specification: &source.SourceAppSpecification{
						App: &armappcontainers.ContainerApp{},
					},
					Err: fmt.Errorf("foobar"),
				},
			},
		}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
			"foo2": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		remoteAppClient.GetSecondResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "sourceApps contains errors, stopping reconciliation")
		actions := remoteAppClient.Actions()
		require.Len(t, actions, 0)
	}()

	// test appCache
	// sourceClient.Get() returns one SourceApp without error
	// first remoteClient.Get() returns one RemoteApp without error
	// second remoteClient.Get() returns one RemoteApp without error
	func() {
		defer resetClients()
		now := time.Now()
		later := time.Now().Add(1 * time.Minute)
		sourceApp1 := source.SourceApp{
			Kind:       "AzureContainerApp",
			APIVersion: "aca.xenit.io/v1alpha2",
			Metadata: map[string]string{
				"name": "foo1",
			},
			Specification: &source.SourceAppSpecification{
				App: &armappcontainers.ContainerApp{
					Name: toPtr("foo1"),
				},
			},
		}
		sourceApp1Updated := source.SourceApp{
			Kind:       "AzureContainerApp",
			APIVersion: "aca.xenit.io/v1alpha2",
			Metadata: map[string]string{
				"name": "foo1",
			},
			Specification: &source.SourceAppSpecification{
				App: &armappcontainers.ContainerApp{
					Name: toPtr("foo1"),
					Tags: map[string]*string{
						"foo": toPtr("bar"),
					},
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
		sourceClient.GetResponse(&source.Sources{
			Apps: &source.SourceApps{
				"foo1": sourceApp1,
			},
		}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{
			"foo1": remoteApp1,
		}, nil)
		remoteAppClient.GetSecondResponse(&remote.RemoteApps{
			"foo1": remoteApp1,
		}, nil)

		// run once and appCache
		{
			err := reconciler.Run(ctx)
			require.NoError(t, err)
			actions := remoteAppClient.Actions()
			require.Len(t, actions, 1)
			require.Equal(t, actions[0].Name, "foo1")
			require.Equal(t, actions[0].Action, remote.InMemAppActionsUpdate)
			remoteAppClient.ResetActions()
		}

		// verify no actions taken
		{
			err := reconciler.Run(ctx)
			require.NoError(t, err)
			actions := remoteAppClient.Actions()
			require.Len(t, actions, 0)
		}

		// verify that update is made if appCache is outdated
		{
			remoteAppClient.GetFirstResponse(&remote.RemoteApps{
				"foo1": remoteApp1Later,
			}, nil)
			remoteAppClient.GetSecondResponse(&remote.RemoteApps{
				"foo1": remoteApp1Later,
			}, nil)
			err := reconciler.Run(ctx)
			require.NoError(t, err)
			actions := remoteAppClient.Actions()
			require.Len(t, actions, 1)
			require.Equal(t, actions[0].Name, "foo1")
			require.Equal(t, actions[0].Action, remote.InMemAppActionsUpdate)
			remoteAppClient.ResetActions()
		}

		// verify that update is made if source app changed
		{
			sourceClient.GetResponse(&source.Sources{
				Apps: &source.SourceApps{
					"foo1": sourceApp1Updated,
				},
			}, defaultFakeRevision, nil)
			remoteAppClient.GetFirstResponse(&remote.RemoteApps{
				"foo1": remoteApp1Later,
			}, nil)
			remoteAppClient.GetSecondResponse(&remote.RemoteApps{
				"foo1": remoteApp1Later,
			}, nil)
			err := reconciler.Run(ctx)
			require.NoError(t, err)
			actions := remoteAppClient.Actions()
			require.Len(t, actions, 1)
			require.Equal(t, actions[0].Name, "foo1")
			require.Equal(t, actions[0].Action, remote.InMemAppActionsUpdate)
			remoteAppClient.ResetActions()
		}
	}()

	// do not delete unmanaged remote apps
	// sourceClient.Get() returns empty SourceApps without error
	// first remoteClient.Get() returns one RemoteApp (non managed) without error
	// second remoteClient.Get() returns one RemoteApp (non managed) without error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{Apps: &source.SourceApps{}}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: false,
			},
		}, nil)
		remoteAppClient.GetSecondResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: false,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.NoError(t, err)
		actions := remoteAppClient.Actions()
		require.Len(t, actions, 0)
	}()

	// delete failure
	// sourceClient.Get() returns empty SourceApps without error
	// first remoteClient.Get() returns one RemoteApp without error
	// remoteClient.Delete() returns error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{Apps: &source.SourceApps{}}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		remoteAppClient.DeleteResponse(fmt.Errorf("delete foobar"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "delete foobar")
	}()

	// update failure
	// sourceClient.Get() returns one SourceApp without error
	// first remoteClient.Get() returns one RemoteApp without error
	// remoteClient.Set() returns error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Apps: &source.SourceApps{
				"foo1": source.SourceApp{
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo1",
					},
					Specification: &source.SourceAppSpecification{
						App: &armappcontainers.ContainerApp{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{
			"foo1": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		remoteAppClient.UpdateResponse(fmt.Errorf("update foobar"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "update foobar")
	}()

	// new app failure
	// sourceClient.Get() returns one SourceApp without error
	// first remoteClient.Get() returns empty RemoteApps without error
	// remoteClient.Set() returns error
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Apps: &source.SourceApps{
				"foo1": source.SourceApp{
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo1",
					},
					Specification: &source.SourceAppSpecification{
						App: &armappcontainers.ContainerApp{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		remoteAppClient.CreateResponse(fmt.Errorf("new app foobar"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "new app foobar")
	}()

	// test remote secret
	func() {
		defer resetClients()
		secretClient.Set("ze-remote-secret", "foobar", time.Now())
		sourceClient.GetResponse(&source.Sources{
			Apps: &source.SourceApps{
				"foo": source.SourceApp{
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &source.SourceAppSpecification{
						App: &armappcontainers.ContainerApp{},
						RemoteSecrets: []source.RemoteSecretSpecification{
							{
								SecretName:       toPtr("ze-app-secret"),
								RemoteSecretName: toPtr("ze-remote-secret"),
							},
						},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		remoteAppClient.GetSecondResponse(&remote.RemoteApps{
			"foo": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.NoError(t, err)
		actions := remoteAppClient.Actions()
		require.Len(t, actions, 1)
		require.Equal(t, "foo", actions[0].Name)
		require.Equal(t, remote.InMemAppActionsCreate, actions[0].Action)
		require.Equal(t, "ze-app-secret", *actions[0].App.Properties.Configuration.Secrets[0].Name)
		require.Equal(t, "foobar", *actions[0].App.Properties.Configuration.Secrets[0].Value)
		cacheValue, ok := secretCache.Get("ze-remote-secret")
		require.True(t, ok)
		require.Equal(t, "foobar", cacheValue)
	}()

	// test remote secret failure
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Apps: &source.SourceApps{
				"foo": source.SourceApp{
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &source.SourceAppSpecification{
						App: &armappcontainers.ContainerApp{},
						RemoteSecrets: []source.RemoteSecretSpecification{
							{
								SecretName:       toPtr("ze-app-secret"),
								RemoteSecretName: toPtr("ze-remote-secret-failure"),
							},
						},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		remoteAppClient.GetSecondResponse(&remote.RemoteApps{
			"foo": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "secret not found \"ze-remote-secret-failure\"")
		actions := remoteAppClient.Actions()
		require.Len(t, actions, 0)
	}()

	// test populate registry
	func() {
		defer resetClients()

		cfg := config.Config{
			ContainerRegistryServer:   "foobar.io",
			ContainerRegistryUsername: "foo",
			ContainerRegistryPassword: "bar",
		}
		reconciler, err := NewReconciler(cfg, sourceClient, remoteAppClient, remoteJobClient, secretClient, notificationClient, metricsClient, appCache, jobCache, secretCache)
		require.NoError(t, err)
		sourceClient.GetResponse(&source.Sources{
			Apps: &source.SourceApps{
				"foo": source.SourceApp{
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &source.SourceAppSpecification{
						App: &armappcontainers.ContainerApp{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		remoteAppClient.GetSecondResponse(&remote.RemoteApps{
			"foo": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		err = reconciler.Run(ctx)
		require.NoError(t, err)
		actions := remoteAppClient.Actions()
		require.Len(t, actions, 1)
		require.Equal(t, "foo", actions[0].Name)
		require.Equal(t, remote.InMemAppActionsCreate, actions[0].Action)
		require.Len(t, actions[0].App.Properties.Configuration.Secrets, 1)
		require.Equal(t, "azcagit-reg-cred", *actions[0].App.Properties.Configuration.Secrets[0].Name)
		require.Equal(t, "bar", *actions[0].App.Properties.Configuration.Secrets[0].Value)
		require.Len(t, actions[0].App.Properties.Configuration.Registries, 1)
		require.Equal(t, "foobar.io", *actions[0].App.Properties.Configuration.Registries[0].Server)
		require.Equal(t, "foo", *actions[0].App.Properties.Configuration.Registries[0].Username)
		require.Equal(t, "azcagit-reg-cred", *actions[0].App.Properties.Configuration.Registries[0].PasswordSecretRef)
	}()

	// test notification success event
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Apps: &source.SourceApps{
				"foo": source.SourceApp{
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &source.SourceAppSpecification{
						App: &armappcontainers.ContainerApp{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		remoteAppClient.GetSecondResponse(&remote.RemoteApps{
			"foo": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.NoError(t, err)
		actions := remoteAppClient.Actions()
		require.Len(t, actions, 1)
		require.Equal(t, actions[0].Name, "foo")
		require.Equal(t, actions[0].Action, remote.InMemAppActionsCreate)
		notifications := notificationClient.GetNotifications()
		require.Len(t, notifications, 1)
		require.Equal(t, notification.NotificationStateSuccess, notifications[0].State)
		require.Equal(t, defaultFakeRevision, notifications[0].Revision)
	}()

	// test notification failure event
	func() {
		defer resetClients()
		sourceClient.GetResponse(nil, defaultFakeRevision, fmt.Errorf("fake unable to parse"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "fake unable to parse")
		notifications := notificationClient.GetNotifications()
		require.Len(t, notifications, 1)
		require.Equal(t, notification.NotificationStateFailure, notifications[0].State)
	}()

	// test two notifications
	func() {
		defer resetClients()
		{
			sourceClient.GetResponse(nil, defaultFakeRevision, fmt.Errorf("ze-failure-one"))
			err := reconciler.Run(ctx)
			require.ErrorContains(t, err, "ze-failure-one")
		}
		{
			sourceClient.GetResponse(nil, defaultFakeRevision, fmt.Errorf("ze-failure-two"))
			err := reconciler.Run(ctx)
			require.ErrorContains(t, err, "ze-failure-two")
		}

		notifications := notificationClient.GetNotifications()
		require.Len(t, notifications, 2)
		require.Equal(t, notification.NotificationStateFailure, notifications[0].State)
		require.Contains(t, notifications[0].Description, "ze-failure-one")
		require.Equal(t, notification.NotificationStateFailure, notifications[1].State)
		require.Contains(t, notifications[1].Description, "ze-failure-two")
	}()

	// test notification with different revisions
	func() {
		defer resetClients()
		{
			sourceClient.GetResponse(&source.Sources{
				Apps: &source.SourceApps{
					"foo": source.SourceApp{
						Kind:       "AzureContainerApp",
						APIVersion: "aca.xenit.io/v1alpha2",
						Metadata: map[string]string{
							"name": "foo",
						},
						Specification: &source.SourceAppSpecification{
							App: &armappcontainers.ContainerApp{},
						},
					},
				},
			}, "first-revision", nil)
			remoteAppClient.GetFirstResponse(&remote.RemoteApps{}, nil)
			remoteAppClient.GetSecondResponse(&remote.RemoteApps{
				"foo": remote.RemoteApp{
					App:     &armappcontainers.ContainerApp{},
					Managed: true,
				},
			}, nil)
			err := reconciler.Run(ctx)
			require.NoError(t, err)

		}
		{
			sourceClient.GetResponse(&source.Sources{
				Apps: &source.SourceApps{
					"foo": source.SourceApp{
						Kind:       "AzureContainerApp",
						APIVersion: "aca.xenit.io/v1alpha2",
						Metadata: map[string]string{
							"name": "foo",
						},
						Specification: &source.SourceAppSpecification{
							App: &armappcontainers.ContainerApp{},
						},
					},
				},
			}, "second-revision", nil)
			remoteAppClient.GetFirstResponse(&remote.RemoteApps{}, nil)
			remoteAppClient.GetSecondResponse(&remote.RemoteApps{
				"foo": remote.RemoteApp{
					App:     &armappcontainers.ContainerApp{},
					Managed: true,
				},
			}, nil)
			err := reconciler.Run(ctx)
			require.NoError(t, err)
		}

		notifications := notificationClient.GetNotifications()
		require.Len(t, notifications, 2)
		require.Equal(t, "first-revision", notifications[0].Revision)
		require.Equal(t, "second-revision", notifications[1].Revision)
	}()

	// test notification deduplication
	func() {
		defer resetClients()
		{
			sourceClient.GetResponse(nil, defaultFakeRevision, fmt.Errorf("ze-failure"))
			err := reconciler.Run(ctx)
			require.ErrorContains(t, err, "ze-failure")
		}
		{
			sourceClient.GetResponse(nil, defaultFakeRevision, fmt.Errorf("ze-failure"))
			err := reconciler.Run(ctx)
			require.ErrorContains(t, err, "ze-failure")
		}

		notifications := notificationClient.GetNotifications()
		require.Len(t, notifications, 1)
		require.Equal(t, notification.NotificationStateFailure, notifications[0].State)
		require.Contains(t, notifications[0].Description, "ze-failure")
	}()

	// test notification error
	func() {
		defer resetClients()

		sourceClient.GetResponse(&source.Sources{
			Apps: &source.SourceApps{
				"foo": source.SourceApp{
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &source.SourceAppSpecification{
						App: &armappcontainers.ContainerApp{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		remoteAppClient.GetSecondResponse(&remote.RemoteApps{
			"foo": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		notificationClient.SendResponse(fmt.Errorf("fake notification error"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "fake notification error")
	}()

	// test locationFilter
	func() {
		defer resetClients()

		cfg := config.Config{
			Location: "foobar",
		}
		reconciler, err := NewReconciler(cfg, sourceClient, remoteAppClient, remoteJobClient, secretClient, notificationClient, metricsClient, appCache, jobCache, secretCache)
		require.NoError(t, err)

		sourceClient.GetResponse(&source.Sources{
			Apps: &source.SourceApps{
				"foo": source.SourceApp{
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &source.SourceAppSpecification{
						LocationFilter: []source.LocationFilterSpecification{
							"zefakeregion",
						},
						App: &armappcontainers.ContainerApp{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		remoteAppClient.GetSecondResponse(&remote.RemoteApps{}, nil)
		err = reconciler.Run(ctx)
		require.NoError(t, err)
		actions := remoteAppClient.Actions()
		require.Len(t, actions, 0)
	}()

	// verify that metrics work
	func() {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Apps: &source.SourceApps{
				"foo": source.SourceApp{
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &source.SourceAppSpecification{
						App: &armappcontainers.ContainerApp{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		remoteAppClient.GetSecondResponse(&remote.RemoteApps{
			"foo": remote.RemoteApp{
				App:     &armappcontainers.ContainerApp{},
				Managed: true,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.NoError(t, err)
		actions := remoteAppClient.Actions()
		require.Len(t, actions, 1)
		require.Equal(t, actions[0].Name, "foo")
		require.Equal(t, actions[0].Action, remote.InMemAppActionsCreate)
		intStats := metricsClient.IntStats()
		require.Len(t, intStats, 1)
		require.Equal(t, 1, intStats[0])
		durationStats := metricsClient.DurationStats()
		require.Len(t, durationStats, 1)
		require.Greater(t, durationStats[0].Nanoseconds(), int64(100))
		successStats := metricsClient.SuccessStats()
		require.Len(t, successStats, 1)
		require.True(t, successStats[0])
	}()
}

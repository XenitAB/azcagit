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
		remoteJobClient.GetFirstResponse(nil, nil)
		remoteJobClient.GetSecondResponse(nil, nil)
		remoteJobClient.ResetGetSecond()
		remoteJobClient.CreateResponse(nil)
		remoteJobClient.UpdateResponse(nil)
		remoteJobClient.DeleteResponse(nil)
		remoteJobClient.ResetActions()
		secretClient.Reset()
		notificationClient.SendResponse(nil)
		notificationClient.ResetNotifications()
		metricsClient.Reset()
		reconciler.previousNotificationEvent = notification.NotificationEvent{}
	}

	t.Run("everything is nil", func(t *testing.T) {
		defer resetClients()
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "sources is nil")
	})

	t.Run("sourceClient.Get() returns error", func(t *testing.T) {
		defer resetClients()
		sourceClient.GetResponse(nil, defaultFakeRevision, fmt.Errorf("foobar"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "failed to get sourceApps: foobar")
	})

	t.Run("sourceClient.Get() returns empty SourceApps without error, sourceClient.Get() returns empty SourceApps without error", func(t *testing.T) {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{Apps: &source.SourceApps{}}, defaultFakeRevision, nil)
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "remoteApps is nil")
	})

	t.Run("sourceClient.Get() returns empty SourceApps without error, first remoteClient.Get() returns error", func(t *testing.T) {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{Apps: &source.SourceApps{}}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(nil, fmt.Errorf("foobar"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "failed to get remoteApps: foobar")
	})

	t.Run("sourceClient.Get() returns empty SourceApps without error, first remoteClient.Get() returns empty RemoteApps without error", func(t *testing.T) {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{Apps: &source.SourceApps{}}, defaultFakeRevision, nil)
		remoteAppClient.GetFirstResponse(&remote.RemoteApps{}, nil)
		err := reconciler.Run(ctx)
		require.NoError(t, err)
	})

	t.Run("sourceClient.Get() returns one SourceApp without error, first remoteClient.Get() returns empty RemoteApps without error, second remoteClient.Get() returns nil RemoteApps with error", func(t *testing.T) {
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
	})

	t.Run("sourceClient.Get() returns one SourceJob without error, first remoteClient.Get() returns empty RemoteJobs without error, second remoteClient.Get() returns nil RemoteJobs with error", func(t *testing.T) {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Jobs: &source.SourceJobs{
				"foo": source.SourceJob{
					Kind:       "AzureContainerJob",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &source.SourceJobSpecification{
						Job: &armappcontainers.Job{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteJobClient.GetFirstResponse(&remote.RemoteJobs{}, nil)
		remoteJobClient.GetSecondResponse(nil, fmt.Errorf("foobar second"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "failed to get new remoteJobs: foobar second")
		actions := remoteJobClient.Actions()
		require.Len(t, actions, 1)
		require.Equal(t, actions[0].Name, "foo")
		require.Equal(t, actions[0].Action, remote.InMemJobActionsCreate)
	})

	t.Run("sourceClient.Get() returns one SourceApp without error, first remoteClient.Get() returns empty RemoteApps without error, second remoteClient.Get() returns one RemoteApp without error", func(t *testing.T) {
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
	})

	t.Run("sourceClient.Get() returns one SourceJob without error, first remoteClient.Get() returns empty RemoteJobs without error, second remoteClient.Get() returns one RemoteJob without error", func(t *testing.T) {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Jobs: &source.SourceJobs{
				"foo": source.SourceJob{
					Kind:       "AzureContainerJob",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &source.SourceJobSpecification{
						Job: &armappcontainers.Job{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteJobClient.GetFirstResponse(&remote.RemoteJobs{}, nil)
		remoteJobClient.GetSecondResponse(&remote.RemoteJobs{
			"foo": remote.RemoteJob{
				Job:     &armappcontainers.Job{},
				Managed: true,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.NoError(t, err)
		actions := remoteJobClient.Actions()
		require.Len(t, actions, 1)
		require.Equal(t, actions[0].Name, "foo")
		require.Equal(t, actions[0].Action, remote.InMemJobActionsCreate)
	})

	t.Run("sourceClient.Get() returns two SourceApps without error, first remoteClient.Get() returns empty RemoteApps without error, second remoteClient.Get() returns one RemoteApp without error", func(t *testing.T) {
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
		require.ErrorContains(t, err, "unable to locate app foo2 after create or update")
		actions := remoteAppClient.Actions()
		require.Len(t, actions, 2)
		require.Equal(t, actions[0].Name, "foo1")
		require.Equal(t, actions[0].Action, remote.InMemAppActionsCreate)
		require.Equal(t, actions[1].Name, "foo2")
		require.Equal(t, actions[1].Action, remote.InMemAppActionsCreate)
	})

	t.Run("sourceClient.Get() returns two SourceJobs without error, first remoteClient.Get() returns empty RemoteJobs without error, second remoteClient.Get() returns one RemoteJob without error", func(t *testing.T) {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Jobs: &source.SourceJobs{
				"foo1": source.SourceJob{
					Kind:       "AzureContainerJob",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo1",
					},
					Specification: &source.SourceJobSpecification{
						Job: &armappcontainers.Job{},
					},
				},
				"foo2": source.SourceJob{
					Kind:       "AzureContainerJob",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo2",
					},
					Specification: &source.SourceJobSpecification{
						Job: &armappcontainers.Job{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteJobClient.GetFirstResponse(&remote.RemoteJobs{}, nil)
		remoteJobClient.GetSecondResponse(&remote.RemoteJobs{
			"foo1": remote.RemoteJob{
				Job:     &armappcontainers.Job{},
				Managed: true,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "unable to locate job foo2 after create or update")
		actions := remoteJobClient.Actions()
		require.Len(t, actions, 2)
		require.Equal(t, actions[0].Name, "foo1")
		require.Equal(t, actions[0].Action, remote.InMemJobActionsCreate)
		require.Equal(t, actions[1].Name, "foo2")
		require.Equal(t, actions[1].Action, remote.InMemJobActionsCreate)
	})

	t.Run("sourceClient.Get() returns two SourceApps without error, first remoteClient.Get() returns empty RemoteApps without error, second remoteClient.Get() returns two RemoteApps without error", func(t *testing.T) {
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
	})

	t.Run("sourceClient.Get() returns two SourceJobs without error, first remoteClient.Get() returns empty RemoteJobs without error, second remoteClient.Get() returns two RemoteJobs without error", func(t *testing.T) {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Jobs: &source.SourceJobs{
				"foo1": source.SourceJob{
					Kind:       "AzureContainerJob",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo1",
					},
					Specification: &source.SourceJobSpecification{
						Job: &armappcontainers.Job{},
					},
				},
				"foo2": source.SourceJob{
					Kind:       "AzureContainerJob",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo2",
					},
					Specification: &source.SourceJobSpecification{
						Job: &armappcontainers.Job{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteJobClient.GetFirstResponse(&remote.RemoteJobs{}, nil)
		remoteJobClient.GetSecondResponse(&remote.RemoteJobs{
			"foo1": remote.RemoteJob{
				Job:     &armappcontainers.Job{},
				Managed: true,
			},
			"foo2": remote.RemoteJob{
				Job:     &armappcontainers.Job{},
				Managed: true,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.NoError(t, err)
		actions := remoteJobClient.Actions()
		require.Len(t, actions, 2)
		require.Equal(t, actions[0].Name, "foo1")
		require.Equal(t, actions[0].Action, remote.InMemJobActionsCreate)
		require.Equal(t, actions[1].Name, "foo2")
		require.Equal(t, actions[1].Action, remote.InMemJobActionsCreate)
	})

	t.Run("sourceClient.Get() returns one SourceApps without error, first remoteClient.Get() returns two RemoteApps without error, second remoteClient.Get() returns one RemoteApps without error", func(t *testing.T) {
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
	})

	t.Run("sourceClient.Get() returns one SourceJobs without error, first remoteClient.Get() returns two RemoteJobs without error, second remoteClient.Get() returns one RemoteJobs without error", func(t *testing.T) {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Jobs: &source.SourceJobs{
				"foo1": source.SourceJob{
					Kind:       "AzureContainerJob",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo1",
					},
					Specification: &source.SourceJobSpecification{
						Job: &armappcontainers.Job{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteJobClient.GetFirstResponse(&remote.RemoteJobs{
			"foo1": remote.RemoteJob{
				Job:     &armappcontainers.Job{},
				Managed: true,
			},
			"foo2": remote.RemoteJob{
				Job:     &armappcontainers.Job{},
				Managed: true,
			},
		}, nil)
		remoteJobClient.GetSecondResponse(&remote.RemoteJobs{
			"foo1": remote.RemoteJob{
				Job:     &armappcontainers.Job{},
				Managed: true,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.NoError(t, err)
		actions := remoteJobClient.Actions()
		require.Len(t, actions, 2)
		require.Equal(t, actions[0].Name, "foo2")
		require.Equal(t, actions[0].Action, remote.InMemJobActionsDelete)
		require.Equal(t, actions[1].Name, "foo1")
		require.Equal(t, actions[1].Action, remote.InMemJobActionsUpdate)
	})

	t.Run("verify that if any sourceApp contains parsing error, reconciliation stops, sourceClient.Get() returns one SourceApp with error, first remoteClient.Get() returns two RemoteApps without error, second remoteClient.Get() returns one RemoteApps without error", func(t *testing.T) {
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
	})

	t.Run("verify that if any sourceJob contains parsing error, reconciliation stops, sourceClient.Get() returns one SourceJob with error, first remoteClient.Get() returns two RemoteJobs without error, second remoteClient.Get() returns one RemoteJobs without error", func(t *testing.T) {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Jobs: &source.SourceJobs{
				"foo1": source.SourceJob{
					Kind:       "AzureContainerJob",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo1",
					},
					Specification: &source.SourceJobSpecification{
						Job: &armappcontainers.Job{},
					},
					Err: fmt.Errorf("foobar"),
				},
			},
		}, defaultFakeRevision, nil)
		remoteJobClient.GetFirstResponse(&remote.RemoteJobs{
			"foo1": remote.RemoteJob{
				Job:     &armappcontainers.Job{},
				Managed: true,
			},
			"foo2": remote.RemoteJob{
				Job:     &armappcontainers.Job{},
				Managed: true,
			},
		}, nil)
		remoteJobClient.GetSecondResponse(&remote.RemoteJobs{
			"foo1": remote.RemoteJob{
				Job:     &armappcontainers.Job{},
				Managed: true,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "sourceJobs contains errors, stopping reconciliation")
		actions := remoteJobClient.Actions()
		require.Len(t, actions, 0)
	})

	t.Run("test appCache, sourceClient.Get() returns one SourceApp without error, first remoteClient.Get() returns one RemoteApp without error, second remoteClient.Get() returns one RemoteApp without error", func(t *testing.T) {
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

		t.Run("run once and appCache", func(t *testing.T) {
			err := reconciler.Run(ctx)
			require.NoError(t, err)
			actions := remoteAppClient.Actions()
			require.Len(t, actions, 1)
			require.Equal(t, actions[0].Name, "foo1")
			require.Equal(t, actions[0].Action, remote.InMemAppActionsUpdate)
			remoteAppClient.ResetActions()
		})

		t.Run("verify no actions taken", func(t *testing.T) {
			err := reconciler.Run(ctx)
			require.NoError(t, err)
			actions := remoteAppClient.Actions()
			require.Len(t, actions, 0)
		})

		t.Run("verify that update is made if appCache is outdated", func(t *testing.T) {
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
		})

		t.Run("verify that update is made if source app changed", func(t *testing.T) {
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
		})
	})

	t.Run("test jobCache, sourceClient.Get() returns one SourceJob without error, first remoteClient.Get() returns one RemoteJob without error, second remoteClient.Get() returns one RemoteJob without error", func(t *testing.T) {
		defer resetClients()
		now := time.Now()
		later := time.Now().Add(1 * time.Minute)
		sourceJob1 := source.SourceJob{
			Kind:       "AzureContainerJob",
			APIVersion: "aca.xenit.io/v1alpha2",
			Metadata: map[string]string{
				"name": "foo1",
			},
			Specification: &source.SourceJobSpecification{
				Job: &armappcontainers.Job{
					Name: toPtr("foo1"),
				},
			},
		}
		sourceJob1Updated := source.SourceJob{
			Kind:       "AzureContainerJob",
			APIVersion: "aca.xenit.io/v1alpha2",
			Metadata: map[string]string{
				"name": "foo1",
			},
			Specification: &source.SourceJobSpecification{
				Job: &armappcontainers.Job{
					Name: toPtr("foo1"),
					Tags: map[string]*string{
						"foo": toPtr("bar"),
					},
				},
			},
		}
		remoteJob1 := remote.RemoteJob{
			Job: &armappcontainers.Job{
				SystemData: &armappcontainers.SystemData{
					LastModifiedAt: &now,
				},
			},
			Managed: true,
		}
		remoteJob1Later := remote.RemoteJob{
			Job: &armappcontainers.Job{
				SystemData: &armappcontainers.SystemData{
					LastModifiedAt: &later,
				},
			},
			Managed: true,
		}
		sourceClient.GetResponse(&source.Sources{
			Jobs: &source.SourceJobs{
				"foo1": sourceJob1,
			},
		}, defaultFakeRevision, nil)
		remoteJobClient.GetFirstResponse(&remote.RemoteJobs{
			"foo1": remoteJob1,
		}, nil)
		remoteJobClient.GetSecondResponse(&remote.RemoteJobs{
			"foo1": remoteJob1,
		}, nil)

		t.Run("run once and jobCache", func(t *testing.T) {
			err := reconciler.Run(ctx)
			require.NoError(t, err)
			actions := remoteJobClient.Actions()
			require.Len(t, actions, 1)
			require.Equal(t, actions[0].Name, "foo1")
			require.Equal(t, actions[0].Action, remote.InMemJobActionsUpdate)
			remoteJobClient.ResetActions()
		})

		t.Run("verify no actions taken", func(t *testing.T) {
			err := reconciler.Run(ctx)
			require.NoError(t, err)
			actions := remoteJobClient.Actions()
			require.Len(t, actions, 0)
		})

		t.Run("verify that update is made if jobCache is outdated", func(t *testing.T) {
			remoteJobClient.GetFirstResponse(&remote.RemoteJobs{
				"foo1": remoteJob1Later,
			}, nil)
			remoteJobClient.GetSecondResponse(&remote.RemoteJobs{
				"foo1": remoteJob1Later,
			}, nil)
			err := reconciler.Run(ctx)
			require.NoError(t, err)
			actions := remoteJobClient.Actions()
			require.Len(t, actions, 1)
			require.Equal(t, actions[0].Name, "foo1")
			require.Equal(t, actions[0].Action, remote.InMemJobActionsUpdate)
			remoteJobClient.ResetActions()
		})

		t.Run("verify that update is made if source app changed", func(t *testing.T) {
			sourceClient.GetResponse(&source.Sources{
				Jobs: &source.SourceJobs{
					"foo1": sourceJob1Updated,
				},
			}, defaultFakeRevision, nil)
			remoteJobClient.GetFirstResponse(&remote.RemoteJobs{
				"foo1": remoteJob1Later,
			}, nil)
			remoteJobClient.GetSecondResponse(&remote.RemoteJobs{
				"foo1": remoteJob1Later,
			}, nil)
			err := reconciler.Run(ctx)
			require.NoError(t, err)
			actions := remoteJobClient.Actions()
			require.Len(t, actions, 1)
			require.Equal(t, actions[0].Name, "foo1")
			require.Equal(t, actions[0].Action, remote.InMemJobActionsUpdate)
			remoteJobClient.ResetActions()
		})
	})

	t.Run("do not delete unmanaged remoteApps, sourceClient.Get() returns empty SourceApps without error, first remoteClient.Get() returns one RemoteApp (non managed) without error, second remoteClient.Get() returns one RemoteApp (non managed) without error", func(t *testing.T) {
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
	})

	t.Run("do not delete unmanaged remoteJobs, sourceClient.Get() returns empty SourceJobs without error, first remoteClient.Get() returns one RemoteJob (non managed) without error, second remoteClient.Get() returns one RemoteJob (non managed) without error", func(t *testing.T) {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{Jobs: &source.SourceJobs{}}, defaultFakeRevision, nil)
		remoteJobClient.GetFirstResponse(&remote.RemoteJobs{
			"foo1": remote.RemoteJob{
				Job:     &armappcontainers.Job{},
				Managed: false,
			},
		}, nil)
		remoteJobClient.GetSecondResponse(&remote.RemoteJobs{
			"foo1": remote.RemoteJob{
				Job:     &armappcontainers.Job{},
				Managed: false,
			},
		}, nil)
		err := reconciler.Run(ctx)
		require.NoError(t, err)
		actions := remoteJobClient.Actions()
		require.Len(t, actions, 0)
	})

	t.Run("delete failure, sourceClient.Get() returns empty SourceApps without error, first remoteClient.Get() returns one RemoteApp without error, remoteClient.Delete() returns error", func(t *testing.T) {
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
	})

	t.Run("delete failure, sourceClient.Get() returns empty SourceJobs without error, first remoteClient.Get() returns one RemoteJob without error, remoteClient.Delete() returns error", func(t *testing.T) {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{Jobs: &source.SourceJobs{}}, defaultFakeRevision, nil)
		remoteJobClient.GetFirstResponse(&remote.RemoteJobs{
			"foo1": remote.RemoteJob{
				Job:     &armappcontainers.Job{},
				Managed: true,
			},
		}, nil)
		remoteJobClient.DeleteResponse(fmt.Errorf("delete foobar"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "delete foobar")
	})

	t.Run("update failure, sourceClient.Get() returns one SourceApp without error, first remoteClient.Get() returns one RemoteApp without error, remoteClient.Set() returns error", func(t *testing.T) {
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
	})

	t.Run("update failure, sourceClient.Get() returns one SourceJob without error, first remoteClient.Get() returns one RemoteJob without error, remoteClient.Set() returns error", func(t *testing.T) {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Jobs: &source.SourceJobs{
				"foo1": source.SourceJob{
					Kind:       "AzureContainerJob",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo1",
					},
					Specification: &source.SourceJobSpecification{
						Job: &armappcontainers.Job{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteJobClient.GetFirstResponse(&remote.RemoteJobs{
			"foo1": remote.RemoteJob{
				Job:     &armappcontainers.Job{},
				Managed: true,
			},
		}, nil)
		remoteJobClient.UpdateResponse(fmt.Errorf("update foobar"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "update foobar")
	})

	t.Run("new app failure, sourceClient.Get() returns one SourceApp without error, first remoteClient.Get() returns empty RemoteApps without error, remoteClient.Set() returns error", func(t *testing.T) {
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
	})

	t.Run("new job failure, sourceClient.Get() returns one SourceJob without error, first remoteClient.Get() returns empty RemoteJobs without error, remoteClient.Set() returns error", func(t *testing.T) {
		defer resetClients()
		sourceClient.GetResponse(&source.Sources{
			Jobs: &source.SourceJobs{
				"foo1": source.SourceJob{
					Kind:       "AzureContainerJob",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo1",
					},
					Specification: &source.SourceJobSpecification{
						Job: &armappcontainers.Job{},
					},
				},
			},
		}, defaultFakeRevision, nil)
		remoteJobClient.GetFirstResponse(&remote.RemoteJobs{}, nil)
		remoteJobClient.CreateResponse(fmt.Errorf("new app foobar"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "new app foobar")
	})

	t.Run("test remote secret", func(t *testing.T) {
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
	})

	t.Run("test remote secret failure", func(t *testing.T) {
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
	})

	t.Run("test populate registry", func(t *testing.T) {
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
	})

	t.Run("test notification success event", func(t *testing.T) {
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
	})

	t.Run("test notification failure event", func(t *testing.T) {
		defer resetClients()
		sourceClient.GetResponse(nil, defaultFakeRevision, fmt.Errorf("fake unable to parse"))
		err := reconciler.Run(ctx)
		require.ErrorContains(t, err, "fake unable to parse")
		notifications := notificationClient.GetNotifications()
		require.Len(t, notifications, 1)
		require.Equal(t, notification.NotificationStateFailure, notifications[0].State)
	})

	t.Run("test two notifications", func(t *testing.T) {
		defer resetClients()
		t.Run("first", func(t *testing.T) {
			sourceClient.GetResponse(nil, defaultFakeRevision, fmt.Errorf("ze-failure-one"))
			err := reconciler.Run(ctx)
			require.ErrorContains(t, err, "ze-failure-one")
		})

		t.Run("second", func(t *testing.T) {
			sourceClient.GetResponse(nil, defaultFakeRevision, fmt.Errorf("ze-failure-two"))
			err := reconciler.Run(ctx)
			require.ErrorContains(t, err, "ze-failure-two")
		})

		notifications := notificationClient.GetNotifications()
		require.Len(t, notifications, 2)
		require.Equal(t, notification.NotificationStateFailure, notifications[0].State)
		require.Contains(t, notifications[0].Description, "ze-failure-one")
		require.Equal(t, notification.NotificationStateFailure, notifications[1].State)
		require.Contains(t, notifications[1].Description, "ze-failure-two")
	})

	t.Run("test notification with different revisions", func(t *testing.T) {
		defer resetClients()
		t.Run("first revision", func(t *testing.T) {
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
		})

		t.Run("second revision", func(t *testing.T) {
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
		})

		notifications := notificationClient.GetNotifications()
		require.Len(t, notifications, 2)
		require.Equal(t, "first-revision", notifications[0].Revision)
		require.Equal(t, "second-revision", notifications[1].Revision)
	})

	t.Run("test notification deduplication", func(t *testing.T) {
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
	})

	t.Run("test notification error", func(t *testing.T) {
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
	})

	t.Run("test locationFilter", func(t *testing.T) {
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
	})

	t.Run("verify that metrics work", func(t *testing.T) {
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
	})
}

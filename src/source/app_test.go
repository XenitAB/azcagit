package source

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
	"github.com/stretchr/testify/require"
	"github.com/xenitab/azcagit/src/config"
)

func TestSourceApp(t *testing.T) {
	cases := []struct {
		testDescription string
		rawYaml         string
		expectedResult  SourceApp
		expectedError   string
	}{
		{
			testDescription: "plain working",
			rawYaml: `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
metadata:
  name: foo
`,
			expectedResult: SourceApp{},
			expectedError:  "spec is missing",
		},
		{
			testDescription: "invalid kind",
			rawYaml: `
kind: foobar
apiVersion: aca.xenit.io/v1alpha1
metadata:
  name: foo
`,
			expectedResult: SourceApp{},
			expectedError:  "kind should be AzureContainerApp",
		},
		{
			testDescription: "invalid apiVersion",
			rawYaml: `
kind: AzureContainerApp
apiVersion: foobar
metadata:
  name: foo
`,
			expectedResult: SourceApp{},
			expectedError:  "apiVersion for AzureContainerApp should be aca.xenit.io/v1alpha1",
		},
		{
			testDescription: "containerapp active revisions mode",
			rawYaml: `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
metadata:
  name: foo
spec:
  app:
    properties:
      configuration:
        activeRevisionsMode: Single
`,
			expectedResult: SourceApp{
				Kind:       "AzureContainerApp",
				APIVersion: "aca.xenit.io/v1alpha1",
				Metadata: map[string]string{
					"name": "foo",
				},
				Specification: &SourceAppSpecification{
					App: &armappcontainers.ContainerApp{
						Properties: &armappcontainers.ContainerAppProperties{
							ManagedEnvironmentID: toPtr("ze-managedEnvironmentID"),
							Configuration: &armappcontainers.Configuration{
								ActiveRevisionsMode: toPtr(armappcontainers.ActiveRevisionsModeSingle),
							},
						},
						Location: toPtr("ze-location"),
						Tags: map[string]*string{
							"aca.xenit.io": toPtr("true"),
						},
					},
				},
			},
			expectedError: "",
		},
		{
			testDescription: "containerapp invalid property",
			rawYaml: `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
metadata:
  name: foo
spec:
  app:
    foobar: baz
`,
			expectedResult: SourceApp{},
			expectedError:  "json: unknown field \"foobar\"",
		},
		{
			testDescription: "containerapp with multiple properties",
			rawYaml: `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
metadata:
  name: foo
spec:
  app:
    properties:
      configuration:
        activeRevisionsMode: Single
      template:
        containers:
        - name: simple-hello-world-container
          image: mcr.microsoft.com/azuredocs/containerapps-helloworld:latest
          resources:
            cpu: 0.25
            memory: .5Gi
        scale:
          minReplicas: 1
          maxReplicas: 1
`,
			expectedResult: SourceApp{
				Kind:       "AzureContainerApp",
				APIVersion: "aca.xenit.io/v1alpha1",
				Metadata: map[string]string{
					"name": "foo",
				},
				Specification: &SourceAppSpecification{
					App: &armappcontainers.ContainerApp{
						Location: toPtr("ze-location"),
						Tags: map[string]*string{
							"aca.xenit.io": toPtr("true"),
						},
						Identity: nil,
						Properties: &armappcontainers.ContainerAppProperties{
							ManagedEnvironmentID: toPtr("ze-managedEnvironmentID"),
							Configuration: &armappcontainers.Configuration{
								ActiveRevisionsMode: toPtr(armappcontainers.ActiveRevisionsModeSingle),
							},
							Template: &armappcontainers.Template{
								Containers: []*armappcontainers.Container{
									{
										Name:  toPtr("simple-hello-world-container"),
										Image: toPtr("mcr.microsoft.com/azuredocs/containerapps-helloworld:latest"),
										Resources: &armappcontainers.ContainerResources{
											CPU:    toPtr(float64(0.25)),
											Memory: toPtr(".5Gi"),
										},
									},
								},
								Scale: &armappcontainers.Scale{
									MaxReplicas: toPtr(int32(1)),
									MinReplicas: toPtr(int32(1)),
								},
							},
						},
					},
				},
			},
			expectedError: "",
		},
	}

	for i, c := range cases {
		t.Logf("Test #%d: %s", i, c.testDescription)
		app := SourceApp{}
		err := app.Unmarshal([]byte(c.rawYaml), config.Config{
			Location:             "ze-location",
			ManagedEnvironmentID: "ze-managedEnvironmentID",
		})
		if c.expectedError != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), c.expectedError)
		} else {
			require.NoError(t, err)
		}
		require.Equal(t, c.expectedResult, app)
	}
}

func TestSourceApps(t *testing.T) {
	cases := []struct {
		testDescription string
		rawYaml         string
		expectedResult  SourceApps
		expectedLenght  int
		expectedError   string
	}{
		{
			testDescription: "plain working, single document",
			rawYaml: `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
metadata:
 name: foo
spec:
  app:
    properties:
      configuration:
        activeRevisionsMode: Single
`,
			expectedResult: SourceApps{
				"foo": {
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha1",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &SourceAppSpecification{
						App: &armappcontainers.ContainerApp{
							Properties: &armappcontainers.ContainerAppProperties{
								ManagedEnvironmentID: toPtr("ze-managedEnvironmentID"),
								Configuration: &armappcontainers.Configuration{
									ActiveRevisionsMode: toPtr(armappcontainers.ActiveRevisionsModeSingle),
								},
							},
							Location: toPtr("ze-location"),
							Tags: map[string]*string{
								"aca.xenit.io": toPtr("true"),
							},
						},
					},
				},
			},
			expectedLenght: 1,
			expectedError:  "",
		},
		{
			testDescription: "plain working, two documents",
			rawYaml: `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
metadata:
 name: foo
spec:
  app:
    properties:
      configuration:
        activeRevisionsMode: Single
---
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
metadata:
 name: bar
spec:
  app:
    properties:
      configuration:
        activeRevisionsMode: Single
`,
			expectedResult: SourceApps{
				"foo": {
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha1",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &SourceAppSpecification{
						App: &armappcontainers.ContainerApp{
							Properties: &armappcontainers.ContainerAppProperties{
								ManagedEnvironmentID: toPtr("ze-managedEnvironmentID"),
								Configuration: &armappcontainers.Configuration{
									ActiveRevisionsMode: toPtr(armappcontainers.ActiveRevisionsModeSingle),
								},
							},
							Location: toPtr("ze-location"),
							Tags: map[string]*string{
								"aca.xenit.io": toPtr("true"),
							},
						},
					},
				},
				"bar": {
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha1",
					Metadata: map[string]string{
						"name": "bar",
					},
					Specification: &SourceAppSpecification{
						App: &armappcontainers.ContainerApp{
							Properties: &armappcontainers.ContainerAppProperties{
								ManagedEnvironmentID: toPtr("ze-managedEnvironmentID"),
								Configuration: &armappcontainers.Configuration{
									ActiveRevisionsMode: toPtr(armappcontainers.ActiveRevisionsModeSingle),
								},
							},
							Location: toPtr("ze-location"),
							Tags: map[string]*string{
								"aca.xenit.io": toPtr("true"),
							},
						},
					},
				},
			},
			expectedLenght: 2,
			expectedError:  "",
		},
		{
			testDescription: "one valid, one invalid",
			rawYaml: `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
metadata:
 name: foo
spec:
  app:
    properties:
      configuration:
        activeRevisionsMode: Single
---
kind: foobar
apiVersion: aca.xenit.io/v1alpha1
metadata:
 name: bar
spec:
  app:
    properties:
      configuration:
        activeRevisionsMode: Single
`,
			expectedResult: SourceApps{
				"foo": {
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha1",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &SourceAppSpecification{
						App: &armappcontainers.ContainerApp{
							Properties: &armappcontainers.ContainerAppProperties{
								ManagedEnvironmentID: toPtr("ze-managedEnvironmentID"),
								Configuration: &armappcontainers.Configuration{
									ActiveRevisionsMode: toPtr(armappcontainers.ActiveRevisionsModeSingle),
								},
							},
							Location: toPtr("ze-location"),
							Tags: map[string]*string{
								"aca.xenit.io": toPtr("true"),
							},
						},
					},
				},
			},
			expectedLenght: 2,
			expectedError:  "kind should be AzureContainerApp",
		},
	}

	for i, c := range cases {
		t.Logf("Test #%d: %s", i, c.testDescription)
		apps := SourceApps{}
		apps.Unmarshal("foobar/baz.yaml", []byte(c.rawYaml), config.Config{
			Location:             "ze-location",
			ManagedEnvironmentID: "ze-managedEnvironmentID",
		})
		require.Len(t, apps, c.expectedLenght)
		if c.expectedError != "" {
			require.ErrorContains(t, apps.Error(), c.expectedError)
		} else {
			require.NoError(t, apps.Error())
		}

		appsWithoutErrors := SourceApps{}
		for name, app := range apps {
			if app.Error() != nil {
				continue
			}
			appsWithoutErrors[name] = app

		}
		require.Equal(t, c.expectedResult, appsWithoutErrors)
	}
}

func TestSourceAppSetSecret(t *testing.T) {
	// fails with app is nil
	{
		app := SourceApp{}
		err := app.SetSecret("foo", "bar")
		require.ErrorContains(t, err, "app is nil")
	}
	{
		app := SourceApp{
			Specification: &SourceAppSpecification{},
		}
		err := app.SetSecret("foo", "bar")
		require.ErrorContains(t, err, "app is nil")
	}

	// fails with secret already exists
	{
		app := SourceApp{
			Specification: &SourceAppSpecification{
				App: &armappcontainers.ContainerApp{
					Properties: &armappcontainers.ContainerAppProperties{
						Configuration: &armappcontainers.Configuration{
							Secrets: []*armappcontainers.Secret{
								{
									Name:  toPtr("foo"),
									Value: toPtr("bar"),
								},
							},
						},
					},
				},
			},
		}
		err := app.SetSecret("foo", "bar")
		require.ErrorContains(t, err, "a secret with name \"foo\" already exists")
	}

	// working with no secrets
	{
		app := SourceApp{
			Specification: &SourceAppSpecification{
				App: &armappcontainers.ContainerApp{},
			},
		}
		err := app.SetSecret("foo", "bar")
		require.NoError(t, err)
		require.Len(t, app.Specification.App.Properties.Configuration.Secrets, 1)
		require.Equal(t, "foo", *app.Specification.App.Properties.Configuration.Secrets[0].Name)
		require.Equal(t, "bar", *app.Specification.App.Properties.Configuration.Secrets[0].Value)
	}

	// working with other secrets
	{
		app := SourceApp{
			Specification: &SourceAppSpecification{
				App: &armappcontainers.ContainerApp{
					Properties: &armappcontainers.ContainerAppProperties{
						Configuration: &armappcontainers.Configuration{
							Secrets: []*armappcontainers.Secret{
								{
									Name:  toPtr("baz"),
									Value: toPtr("foobar"),
								},
							},
						},
					},
				},
			},
		}
		err := app.SetSecret("foo", "bar")
		require.NoError(t, err)
		require.Len(t, app.Specification.App.Properties.Configuration.Secrets, 2)
		require.Equal(t, "baz", *app.Specification.App.Properties.Configuration.Secrets[0].Name)
		require.Equal(t, "foobar", *app.Specification.App.Properties.Configuration.Secrets[0].Value)
		require.Equal(t, "foo", *app.Specification.App.Properties.Configuration.Secrets[1].Name)
		require.Equal(t, "bar", *app.Specification.App.Properties.Configuration.Secrets[1].Value)
	}

	// working with SourceApps
	{
		app := SourceApp{
			Specification: &SourceAppSpecification{
				App: &armappcontainers.ContainerApp{},
			},
		}
		apps := make(SourceApps)
		apps["foo"] = app

		err := apps.SetAppSecret("foo", "bar", "baz")
		require.NoError(t, err)

		updatedApp, ok := apps["foo"]
		require.True(t, ok)
		require.Equal(t, "bar", *updatedApp.Specification.App.Properties.Configuration.Secrets[0].Name)
		require.Equal(t, "baz", *updatedApp.Specification.App.Properties.Configuration.Secrets[0].Value)
	}
}

func TestSourceAppSetRegistry(t *testing.T) {
	// fails with app is nil
	{
		app := SourceApp{}
		err := app.SetRegistry("ze-server", "foo", "bar")
		require.ErrorContains(t, err, "app is nil")
	}
	{
		app := SourceApp{
			Specification: &SourceAppSpecification{},
		}
		err := app.SetRegistry("ze-server", "foo", "bar")
		require.ErrorContains(t, err, "app is nil")
	}

	// fails with secret already exists
	{
		app := SourceApp{
			Specification: &SourceAppSpecification{
				App: &armappcontainers.ContainerApp{
					Properties: &armappcontainers.ContainerAppProperties{
						Configuration: &armappcontainers.Configuration{
							Registries: []*armappcontainers.RegistryCredentials{
								{
									Server:            toPtr("ze-server"),
									Username:          toPtr(""),
									PasswordSecretRef: toPtr(""),
								},
							},
						},
					},
				},
			},
		}
		err := app.SetRegistry("ze-server", "foo", "bar")
		require.ErrorContains(t, err, "the server \"ze-server\" already exists")
	}

	// working with no secrets
	{
		app := SourceApp{
			Specification: &SourceAppSpecification{
				App: &armappcontainers.ContainerApp{},
			},
		}
		err := app.SetRegistry("ze-server", "foo", "bar")
		require.NoError(t, err)
		require.Len(t, app.Specification.App.Properties.Configuration.Secrets, 1)
		require.Len(t, app.Specification.App.Properties.Configuration.Registries, 1)
		require.Equal(t, "azcagit-reg-cred", *app.Specification.App.Properties.Configuration.Secrets[0].Name)
		require.Equal(t, "bar", *app.Specification.App.Properties.Configuration.Secrets[0].Value)
		require.Equal(t, "azcagit-reg-cred", *app.Specification.App.Properties.Configuration.Registries[0].PasswordSecretRef)
		require.Equal(t, "ze-server", *app.Specification.App.Properties.Configuration.Registries[0].Server)
		require.Equal(t, "foo", *app.Specification.App.Properties.Configuration.Registries[0].Username)
	}

	// working with other secrets
	{
		app := SourceApp{
			Specification: &SourceAppSpecification{
				App: &armappcontainers.ContainerApp{
					Properties: &armappcontainers.ContainerAppProperties{
						Configuration: &armappcontainers.Configuration{
							Registries: []*armappcontainers.RegistryCredentials{
								{
									Server:            toPtr("ze-other-server"),
									Username:          toPtr("ze-other-username"),
									PasswordSecretRef: toPtr("ze-other-password-ref"),
								},
							},
						},
					},
				},
			},
		}
		err := app.SetRegistry("ze-server", "foo", "bar")
		require.NoError(t, err)
		require.Len(t, app.Specification.App.Properties.Configuration.Secrets, 1)
		require.Len(t, app.Specification.App.Properties.Configuration.Registries, 2)
		require.Equal(t, "azcagit-reg-cred", *app.Specification.App.Properties.Configuration.Secrets[0].Name)
		require.Equal(t, "bar", *app.Specification.App.Properties.Configuration.Secrets[0].Value)
		require.Equal(t, "ze-other-password-ref", *app.Specification.App.Properties.Configuration.Registries[0].PasswordSecretRef)
		require.Equal(t, "ze-other-server", *app.Specification.App.Properties.Configuration.Registries[0].Server)
		require.Equal(t, "ze-other-username", *app.Specification.App.Properties.Configuration.Registries[0].Username)
		require.Equal(t, "azcagit-reg-cred", *app.Specification.App.Properties.Configuration.Registries[1].PasswordSecretRef)
		require.Equal(t, "ze-server", *app.Specification.App.Properties.Configuration.Registries[1].Server)
		require.Equal(t, "foo", *app.Specification.App.Properties.Configuration.Registries[1].Username)
	}

	// working with SourceApps
	{
		app := SourceApp{
			Specification: &SourceAppSpecification{
				App: &armappcontainers.ContainerApp{},
			},
		}
		apps := make(SourceApps)
		apps["foo"] = app

		err := apps.SetAppRegistry("foo", "ze-server", "foo", "bar")
		require.NoError(t, err)

		updatedApp, ok := apps["foo"]
		require.True(t, ok)
		require.Len(t, updatedApp.Specification.App.Properties.Configuration.Secrets, 1)
		require.Len(t, updatedApp.Specification.App.Properties.Configuration.Registries, 1)
		require.Equal(t, "azcagit-reg-cred", *updatedApp.Specification.App.Properties.Configuration.Secrets[0].Name)
		require.Equal(t, "bar", *updatedApp.Specification.App.Properties.Configuration.Secrets[0].Value)
		require.Equal(t, "azcagit-reg-cred", *updatedApp.Specification.App.Properties.Configuration.Registries[0].PasswordSecretRef)
		require.Equal(t, "ze-server", *updatedApp.Specification.App.Properties.Configuration.Registries[0].Server)
		require.Equal(t, "foo", *updatedApp.Specification.App.Properties.Configuration.Registries[0].Username)
	}
}

func TestSourceAppsGetRemoteSecret(t *testing.T) {
	cases := []struct {
		testDescription string
		input           *SourceApps
		expectedOutput  []string
	}{
		{
			testDescription: "single secret",
			input: &SourceApps{
				"foo": {
					Specification: &SourceAppSpecification{
						RemoteSecrets: []RemoteSecretSpecification{
							{
								AppSecretName:    toPtr("foo"),
								RemoteSecretName: toPtr("bar"),
							},
						},
					},
				},
				"bar": {
					Specification: &SourceAppSpecification{
						RemoteSecrets: []RemoteSecretSpecification{
							{
								AppSecretName:    toPtr("baz"),
								RemoteSecretName: toPtr("foobar"),
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"bar",
				"foobar",
			},
		},
		{
			testDescription: "two secrets, same names",
			input: &SourceApps{
				"foo": {
					Specification: &SourceAppSpecification{
						RemoteSecrets: []RemoteSecretSpecification{
							{
								AppSecretName:    toPtr("foo"),
								RemoteSecretName: toPtr("bar"),
							},
						},
					},
				},
				"bar": {
					Specification: &SourceAppSpecification{
						RemoteSecrets: []RemoteSecretSpecification{
							{
								AppSecretName:    toPtr("foo"),
								RemoteSecretName: toPtr("bar"),
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"bar",
			},
		},
	}

	for i, c := range cases {
		t.Logf("Test #%d: %s", i, c.testDescription)
		remoteSecrets := c.input.GetUniqueRemoteSecretNames()
		require.Equal(t, c.expectedOutput, remoteSecrets)
	}
}

func TestParseLocationFilterSpecification(t *testing.T) {
	cases := []struct {
		testDescription        string
		input                  string
		expectedLocationFilter []LocationFilterSpecification
		expectedErrorContains  string
	}{
		{
			testDescription: "no locationFilter specified",
			input: `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
metadata:
  name: foo
spec:
  app:
    properties:
      configuration:
        activeRevisionsMode: Single`,
			expectedLocationFilter: nil,
			expectedErrorContains:  "",
		},
		{
			testDescription: "one locationFilter specified",
			input: `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
metadata:
  name: foo
spec:
  locationFilter:
    - Foo Bar
  app:
    properties:
      configuration:
        activeRevisionsMode: Single`,
			expectedLocationFilter: []LocationFilterSpecification{
				LocationFilterSpecification("foobar"),
			},
			expectedErrorContains: "",
		},
	}

	for i, c := range cases {
		t.Logf("Test #%d: %s", i, c.testDescription)
		app := &SourceApp{}
		err := app.Unmarshal([]byte(c.input), config.Config{
			ManagedEnvironmentID: "ze-me-id",
			Location:             "zefakeregion",
		})
		if c.expectedErrorContains == "" {
			require.NoError(t, err)
		} else {
			require.ErrorContains(t, err, c.expectedErrorContains)
		}

		require.Equal(t, c.expectedLocationFilter, app.Specification.LocationFilter)
	}
}

func TestSourceAppShouldRunInLocation(t *testing.T) {
	cases := []struct {
		currentLocation string
		app             *SourceApp
		expectedResult  bool
	}{
		{
			currentLocation: "foo",
			app:             nil,
			expectedResult:  true,
		},
		{
			currentLocation: "foo",
			app: &SourceApp{
				Specification: &SourceAppSpecification{
					LocationFilter: nil,
				},
			},
			expectedResult: true,
		},
		{
			currentLocation: "foo",
			app: &SourceApp{
				Specification: &SourceAppSpecification{
					LocationFilter: []LocationFilterSpecification{},
				},
			},
			expectedResult: true,
		},
		{
			currentLocation: "foo",
			app: &SourceApp{
				Specification: &SourceAppSpecification{
					LocationFilter: []LocationFilterSpecification{
						"bar",
					},
				},
			},
			expectedResult: false,
		},
		{
			currentLocation: "foo",
			app: &SourceApp{
				Specification: &SourceAppSpecification{
					LocationFilter: []LocationFilterSpecification{
						"foo",
					},
				},
			},
			expectedResult: true,
		},
		{
			currentLocation: "foo",
			app: &SourceApp{
				Specification: &SourceAppSpecification{
					LocationFilter: []LocationFilterSpecification{
						"bar",
						"foo",
					},
				},
			},
			expectedResult: true,
		},
		{
			currentLocation: "Foo Bar",
			app: &SourceApp{
				Specification: &SourceAppSpecification{
					LocationFilter: []LocationFilterSpecification{
						"foobar",
					},
				},
			},
			expectedResult: true,
		},
	}

	for _, c := range cases {
		result := c.app.ShoudRunInLocation(c.currentLocation)
		require.Equal(t, c.expectedResult, result)
	}
}

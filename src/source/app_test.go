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
			testDescription: "containerapp location",
			rawYaml: `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
metadata:
  name: foo
spec:
  location: foobar
`,
			expectedResult: SourceApp{
				Kind:       "AzureContainerApp",
				APIVersion: "aca.xenit.io/v1alpha1",
				Metadata: map[string]string{
					"name": "foo",
				},
				Specification: &armappcontainers.ContainerApp{
					Location: toPtr("foobar"),
					Tags: map[string]*string{
						"aca.xenit.io": toPtr("true"),
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
  location: foobar
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
				Specification: &armappcontainers.ContainerApp{
					Location: toPtr("foobar"),
					Tags: map[string]*string{
						"aca.xenit.io": toPtr("true"),
					},
					Identity: nil,
					Properties: &armappcontainers.ContainerAppProperties{
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
			expectedError: "",
		},
	}

	for i, c := range cases {
		t.Logf("Test #%d: %s", i, c.testDescription)
		app := SourceApp{}
		err := app.Unmarshal([]byte(c.rawYaml), config.Config{})
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
  location: foobar
`,
			expectedResult: SourceApps{
				"foo": {
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha1",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &armappcontainers.ContainerApp{
						Location: toPtr("foobar"),
						Tags: map[string]*string{
							"aca.xenit.io": toPtr("true"),
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
  location: foobar
---
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
metadata:
 name: bar
spec:
  location: foobar
`,
			expectedResult: SourceApps{
				"foo": {
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha1",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &armappcontainers.ContainerApp{
						Location: toPtr("foobar"),
						Tags: map[string]*string{
							"aca.xenit.io": toPtr("true"),
						},
					},
				},
				"bar": {
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha1",
					Metadata: map[string]string{
						"name": "bar",
					},
					Specification: &armappcontainers.ContainerApp{
						Location: toPtr("foobar"),
						Tags: map[string]*string{
							"aca.xenit.io": toPtr("true"),
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
  location: foobar
---
kind: foobar
apiVersion: aca.xenit.io/v1alpha1
metadata:
 name: bar
spec:
  location: foobar
`,
			expectedResult: SourceApps{
				"foo": {
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha1",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &armappcontainers.ContainerApp{
						Location: toPtr("foobar"),
						Tags: map[string]*string{
							"aca.xenit.io": toPtr("true"),
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
		apps.Unmarshal("foobar/baz.yaml", []byte(c.rawYaml), config.Config{})
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

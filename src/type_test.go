package main

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
	"github.com/stretchr/testify/require"
)

func TestAzureContainerApp(t *testing.T) {
	cases := []struct {
		testDescription string
		rawYaml         string
		expectedResult  AzureContainerApp
		expectedError   string
	}{
		{
			testDescription: "plain working",
			rawYaml: `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
name: foo
`,
			expectedResult: AzureContainerApp{
				Kind:       "AzureContainerApp",
				APIVersion: "aca.xenit.io/v1alpha1",
				Name:       "foo",
			},
			expectedError: "",
		},
		{
			testDescription: "invalid kind",
			rawYaml: `
kind: foobar
apiVersion: aca.xenit.io/v1alpha1
name: foo
`,
			expectedResult: AzureContainerApp{},
			expectedError:  "kind should be AzureContainerApp",
		},
		{
			testDescription: "invalid apiVersion",
			rawYaml: `
kind: AzureContainerApp
apiVersion: foobar
name: foo
`,
			expectedResult: AzureContainerApp{},
			expectedError:  "apiVersion for AzureContainerApp should be aca.xenit.io/v1alpha1",
		},
		{
			testDescription: "containerapp location",
			rawYaml: `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
name: foo
containerapp:
  location: foobar
`,
			expectedResult: AzureContainerApp{
				Kind:       "AzureContainerApp",
				APIVersion: "aca.xenit.io/v1alpha1",
				Name:       "foo",
				ContainerApp: &armappcontainers.ContainerApp{
					Location: toPtr("foobar"),
				},
			},
			expectedError: "",
		},
		{
			testDescription: "containerapp invalid property",
			rawYaml: `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
name: foo
containerapp:
  foobar: baz
`,
			expectedResult: AzureContainerApp{},
			expectedError:  "json: unknown field \"foobar\"",
		},
		{
			testDescription: "containerapp with multiple properties",
			rawYaml: `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
name: foo
containerapp:
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
			expectedResult: AzureContainerApp{
				Kind:       "AzureContainerApp",
				APIVersion: "aca.xenit.io/v1alpha1",
				Name:       "foo",
				ContainerApp: &armappcontainers.ContainerApp{
					Location: toPtr("foobar"),
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
		aca := AzureContainerApp{}
		err := aca.Unmarshal([]byte(c.rawYaml), config{})
		if c.expectedError != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), c.expectedError)
		} else {
			require.NoError(t, err)
		}
		require.Equal(t, c.expectedResult, aca)
	}
}

func TestAzureContainerApps(t *testing.T) {
	cases := []struct {
		testDescription string
		rawYaml         string
		expectedResult  AzureContainerApps
		expectedLenght  int
		expectedError   string
	}{
		{
			testDescription: "plain working, single document",
			rawYaml: `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
name: foo
`,
			expectedResult: AzureContainerApps{
				"foo": {
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha1",
					Name:       "foo",
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
name: foo
---
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha1
name: bar
`,
			expectedResult: AzureContainerApps{
				"foo": {
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha1",
					Name:       "foo",
				},
				"bar": {
					Kind:       "AzureContainerApp",
					APIVersion: "aca.xenit.io/v1alpha1",
					Name:       "bar",
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
name: foo
---
kind: foobar
apiVersion: aca.xenit.io/v1alpha1
name: bar
`,
			expectedResult: AzureContainerApps{},
			expectedLenght: 0,
			expectedError:  "kind should be AzureContainerApp",
		},
	}

	for i, c := range cases {
		t.Logf("Test #%d: %s", i, c.testDescription)
		acas := AzureContainerApps{}
		err := acas.Unmarshal([]byte(c.rawYaml), config{})
		require.Len(t, acas, c.expectedLenght)
		if c.expectedError != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), c.expectedError)
		} else {
			require.NoError(t, err)
		}
		require.Equal(t, c.expectedResult, acas)
	}
}

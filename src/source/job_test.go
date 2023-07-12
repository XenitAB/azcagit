package source

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
	"github.com/stretchr/testify/require"
	"github.com/xenitab/azcagit/src/config"
)

func TestSourceJob(t *testing.T) {
	cases := []struct {
		testDescription string
		rawYaml         string
		expectedResult  SourceJob
		expectedError   string
		isContainerJob  bool
	}{
		{
			testDescription: "plain working",
			rawYaml: `
kind: AzureContainerJob
apiVersion: aca.xenit.io/v1alpha2
metadata:
  name: foo
`,
			expectedResult: SourceJob{},
			expectedError:  "spec is missing",
			isContainerJob: true,
		},
		{
			testDescription: "app skipped",
			rawYaml: `
kind: AzureContainerApp
apiVersion: aca.xenit.io/v1alpha2
metadata:
  name: foo
spec:
  app:
    foo: bar
`,
			expectedResult: SourceJob{},
			expectedError:  "",
			isContainerJob: false,
		},
		{
			testDescription: "invalid kind",
			rawYaml: `
kind: foobar
apiVersion: aca.xenit.io/v1alpha2
metadata:
  name: foo
`,
			expectedResult: SourceJob{},
			expectedError:  "",
			isContainerJob: false,
		},
		{
			testDescription: "missing kind",
			rawYaml: `
apiVersion: aca.xenit.io/v1alpha2
metadata:
  name: foo
`,
			expectedResult: SourceJob{},
			expectedError:  "kind is missing",
			isContainerJob: false,
		},
		{
			testDescription: "invalid apiVersion",
			rawYaml: `
kind: AzureContainerJob
apiVersion: foobar
metadata:
  name: foo
`,
			expectedResult: SourceJob{},
			expectedError:  "apiVersion for AzureContainerJob should be aca.xenit.io/v1alpha2",
			isContainerJob: true,
		},
		{
			testDescription: "containerjob replica timeout",
			rawYaml: `
kind: AzureContainerJob
apiVersion: aca.xenit.io/v1alpha2
metadata:
  name: foo
spec:
  job:
    properties:
      configuration:
        replicaTimeout: 1337
`,
			expectedResult: SourceJob{
				Kind:       "AzureContainerJob",
				APIVersion: "aca.xenit.io/v1alpha2",
				Metadata: map[string]string{
					"name": "foo",
				},
				Specification: &SourceJobSpecification{
					Job: &armappcontainers.Job{
						Properties: &armappcontainers.JobProperties{
							EnvironmentID: toPtr("ze-EnvironmentID"),
							Configuration: &armappcontainers.JobConfiguration{
								ReplicaTimeout: toPtr[int32](1337),
							},
						},
						Location: toPtr("ze-location"),
						Tags: map[string]*string{
							"aca.xenit.io": toPtr("true"),
						},
					},
				},
			},
			expectedError:  "",
			isContainerJob: true,
		},
		{
			testDescription: "containerjob invalid property",
			rawYaml: `
kind: AzureContainerJob
apiVersion: aca.xenit.io/v1alpha2
metadata:
  name: foo
spec:
  job:
    foobar: baz
`,
			expectedResult: SourceJob{
				Kind:       "AzureContainerJob",
				APIVersion: "aca.xenit.io/v1alpha2",
				Metadata: map[string]string{
					"name": "foo",
				},
				Specification: &SourceJobSpecification{
					Job: &armappcontainers.Job{
						Location: toPtr("ze-location"),
						Tags: map[string]*string{
							"aca.xenit.io": toPtr("true"),
						},
						Properties: &armappcontainers.JobProperties{
							EnvironmentID: toPtr("ze-EnvironmentID"),
						},
					},
				},
			},
			expectedError:  "",
			isContainerJob: true,
		},
		{
			testDescription: "containerjob with multiple properties",
			rawYaml: `
kind: AzureContainerJob
apiVersion: aca.xenit.io/v1alpha2
metadata:
  name: foo
spec:
  job:
    properties:
      configuration:
        replicaTimeout: 1337
      template:
        containers:
        - name: simple-hello-world-container
          image: mcr.microsoft.com/azuredocs/containerapps-helloworld:latest
          resources:
            cpu: 0.25
            memory: .5Gi
`,
			expectedResult: SourceJob{
				Kind:       "AzureContainerJob",
				APIVersion: "aca.xenit.io/v1alpha2",
				Metadata: map[string]string{
					"name": "foo",
				},
				Specification: &SourceJobSpecification{
					Job: &armappcontainers.Job{
						Location: toPtr("ze-location"),
						Tags: map[string]*string{
							"aca.xenit.io": toPtr("true"),
						},
						Identity: nil,
						Properties: &armappcontainers.JobProperties{
							EnvironmentID: toPtr("ze-EnvironmentID"),
							Configuration: &armappcontainers.JobConfiguration{
								ReplicaTimeout: toPtr[int32](1337),
							},
							Template: &armappcontainers.JobTemplate{
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
							},
						},
					},
				},
			},
			expectedError:  "",
			isContainerJob: true,
		},
		{
			testDescription: "validate that image replacement works",
			rawYaml: `
kind: AzureContainerJob
apiVersion: aca.xenit.io/v1alpha2
metadata:
  name: foo
spec:
  replacements:
    images:
      - imageName: "mcr.microsoft.com/azuredocs/containerapps-helloworld"
        newImageTag: "v0.1"
  job:
    properties:
      configuration:
        replicaTimeout: 1337
      template:
        containers:
        - name: simple-hello-world-container
          image: mcr.microsoft.com/azuredocs/containerapps-helloworld:latest
          resources:
            cpu: 0.25
            memory: .5Gi
`,
			expectedResult: SourceJob{
				Kind:       "AzureContainerJob",
				APIVersion: "aca.xenit.io/v1alpha2",
				Metadata: map[string]string{
					"name": "foo",
				},
				Specification: &SourceJobSpecification{
					Replacements: &ReplacementsSpecification{
						Images: []ImageReplacementSpecification{
							{
								ImageName:   toPtr("mcr.microsoft.com/azuredocs/containerapps-helloworld"),
								NewImageTag: toPtr("v0.1"),
							},
						},
					},
					Job: &armappcontainers.Job{
						Location: toPtr("ze-location"),
						Tags: map[string]*string{
							"aca.xenit.io": toPtr("true"),
						},
						Identity: nil,
						Properties: &armappcontainers.JobProperties{
							EnvironmentID: toPtr("ze-EnvironmentID"),
							Configuration: &armappcontainers.JobConfiguration{
								ReplicaTimeout: toPtr[int32](1337),
							},
							Template: &armappcontainers.JobTemplate{
								Containers: []*armappcontainers.Container{
									{
										Name:  toPtr("simple-hello-world-container"),
										Image: toPtr("mcr.microsoft.com/azuredocs/containerapps-helloworld:v0.1"),
										Resources: &armappcontainers.ContainerResources{
											CPU:    toPtr(float64(0.25)),
											Memory: toPtr(".5Gi"),
										},
									},
								},
							},
						},
					},
				},
			},
			expectedError:  "",
			isContainerJob: true,
		},
	}

	for i, c := range cases {
		t.Logf("Test #%d: %s", i, c.testDescription)
		job := SourceJob{}
		isContainerJob, err := job.Unmarshal([]byte(c.rawYaml), config.Config{
			Location:             "ze-location",
			ManagedEnvironmentID: "ze-EnvironmentID",
		})
		require.Equal(t, c.isContainerJob, isContainerJob)
		if c.expectedError != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), c.expectedError)
		} else {
			require.NoError(t, err)
		}
		require.Equal(t, c.expectedResult, job)
	}
}

func TestSourceJobs(t *testing.T) {
	cases := []struct {
		testDescription string
		rawYaml         string
		expectedResult  SourceJobs
		expectedLenght  int
		expectedError   string
	}{
		{
			testDescription: "plain working, single document",
			rawYaml: `
kind: AzureContainerJob
apiVersion: aca.xenit.io/v1alpha2
metadata:
 name: foo
spec:
  job:
    properties:
      configuration:
        replicaTimeout: 1337
`,
			expectedResult: SourceJobs{
				"foo": {
					Kind:       "AzureContainerJob",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &SourceJobSpecification{
						Job: &armappcontainers.Job{
							Properties: &armappcontainers.JobProperties{
								EnvironmentID: toPtr("ze-EnvironmentID"),
								Configuration: &armappcontainers.JobConfiguration{
									ReplicaTimeout: toPtr[int32](1337),
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
kind: AzureContainerJob
apiVersion: aca.xenit.io/v1alpha2
metadata:
 name: foo
spec:
  job:
    properties:
      configuration:
        replicaTimeout: 1337
---
kind: AzureContainerJob
apiVersion: aca.xenit.io/v1alpha2
metadata:
 name: bar
spec:
  job:
    properties:
      configuration:
        replicaTimeout: 1337
`,
			expectedResult: SourceJobs{
				"foo": {
					Kind:       "AzureContainerJob",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &SourceJobSpecification{
						Job: &armappcontainers.Job{
							Properties: &armappcontainers.JobProperties{
								EnvironmentID: toPtr("ze-EnvironmentID"),
								Configuration: &armappcontainers.JobConfiguration{
									ReplicaTimeout: toPtr[int32](1337),
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
					Kind:       "AzureContainerJob",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "bar",
					},
					Specification: &SourceJobSpecification{
						Job: &armappcontainers.Job{
							Properties: &armappcontainers.JobProperties{
								EnvironmentID: toPtr("ze-EnvironmentID"),
								Configuration: &armappcontainers.JobConfiguration{
									ReplicaTimeout: toPtr[int32](1337),
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
kind: AzureContainerJob
apiVersion: aca.xenit.io/v1alpha2
metadata:
 name: foo
spec:
  job:
    properties:
      configuration:
        replicaTimeout: 1337
---
kind: AzureContainerJob
apiVersion: foobar
metadata:
 name: bar
spec:
  job:
    properties:
      configuration:
        replicaTimeout: 1337
`,
			expectedResult: SourceJobs{
				"foo": {
					Kind:       "AzureContainerJob",
					APIVersion: "aca.xenit.io/v1alpha2",
					Metadata: map[string]string{
						"name": "foo",
					},
					Specification: &SourceJobSpecification{
						Job: &armappcontainers.Job{
							Properties: &armappcontainers.JobProperties{
								EnvironmentID: toPtr("ze-EnvironmentID"),
								Configuration: &armappcontainers.JobConfiguration{
									ReplicaTimeout: toPtr[int32](1337),
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
			expectedError:  "apiVersion for AzureContainerJob should be",
		},
	}

	for i, c := range cases {
		t.Logf("Test #%d: %s", i, c.testDescription)
		jobs := SourceJobs{}
		jobs.Unmarshal("foobar/baz.yaml", []byte(c.rawYaml), config.Config{
			Location:             "ze-location",
			ManagedEnvironmentID: "ze-EnvironmentID",
		})
		require.Len(t, jobs, c.expectedLenght)
		if c.expectedError != "" {
			require.ErrorContains(t, jobs.Error(), c.expectedError)
		} else {
			require.NoError(t, jobs.Error())
		}

		jobsWithoutErrors := SourceJobs{}
		for name, job := range jobs {
			if job.Error() != nil {
				continue
			}
			jobsWithoutErrors[name] = job

		}
		require.Equal(t, c.expectedResult, jobsWithoutErrors)
	}
}

func TestSourceJobSetSecret(t *testing.T) {
	// fails with job is nil
	{
		job := SourceJob{}
		err := job.SetSecret("foo", "bar")
		require.ErrorContains(t, err, "job is nil")
	}
	{
		job := SourceJob{
			Specification: &SourceJobSpecification{},
		}
		err := job.SetSecret("foo", "bar")
		require.ErrorContains(t, err, "job is nil")
	}

	// fails with secret already exists
	{
		job := SourceJob{
			Specification: &SourceJobSpecification{
				Job: &armappcontainers.Job{
					Properties: &armappcontainers.JobProperties{
						Configuration: &armappcontainers.JobConfiguration{
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
		err := job.SetSecret("foo", "bar")
		require.ErrorContains(t, err, "a secret with name \"foo\" already exists")
	}

	// working with no secrets
	{
		job := SourceJob{
			Specification: &SourceJobSpecification{
				Job: &armappcontainers.Job{},
			},
		}
		err := job.SetSecret("foo", "bar")
		require.NoError(t, err)
		require.Len(t, job.Specification.Job.Properties.Configuration.Secrets, 1)
		require.Equal(t, "foo", *job.Specification.Job.Properties.Configuration.Secrets[0].Name)
		require.Equal(t, "bar", *job.Specification.Job.Properties.Configuration.Secrets[0].Value)
	}

	// working with other secrets
	{
		job := SourceJob{
			Specification: &SourceJobSpecification{
				Job: &armappcontainers.Job{
					Properties: &armappcontainers.JobProperties{
						Configuration: &armappcontainers.JobConfiguration{
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
		err := job.SetSecret("foo", "bar")
		require.NoError(t, err)
		require.Len(t, job.Specification.Job.Properties.Configuration.Secrets, 2)
		require.Equal(t, "baz", *job.Specification.Job.Properties.Configuration.Secrets[0].Name)
		require.Equal(t, "foobar", *job.Specification.Job.Properties.Configuration.Secrets[0].Value)
		require.Equal(t, "foo", *job.Specification.Job.Properties.Configuration.Secrets[1].Name)
		require.Equal(t, "bar", *job.Specification.Job.Properties.Configuration.Secrets[1].Value)
	}

	// working with SourceJobs
	{
		job := SourceJob{
			Specification: &SourceJobSpecification{
				Job: &armappcontainers.Job{},
			},
		}
		jobs := make(SourceJobs)
		jobs["foo"] = job

		err := jobs.SetSecret("foo", "bar", "baz")
		require.NoError(t, err)

		updatedJob, ok := jobs["foo"]
		require.True(t, ok)
		require.Equal(t, "bar", *updatedJob.Specification.Job.Properties.Configuration.Secrets[0].Name)
		require.Equal(t, "baz", *updatedJob.Specification.Job.Properties.Configuration.Secrets[0].Value)
	}
}

func TestSourceJobSetRegistry(t *testing.T) {
	// fails with job is nil
	{
		job := SourceJob{}
		err := job.SetRegistry("ze-server", "foo", "bar")
		require.ErrorContains(t, err, "job is nil")
	}
	{
		job := SourceJob{
			Specification: &SourceJobSpecification{},
		}
		err := job.SetRegistry("ze-server", "foo", "bar")
		require.ErrorContains(t, err, "job is nil")
	}

	// fails with secret already exists
	{
		job := SourceJob{
			Specification: &SourceJobSpecification{
				Job: &armappcontainers.Job{
					Properties: &armappcontainers.JobProperties{
						Configuration: &armappcontainers.JobConfiguration{
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
		err := job.SetRegistry("ze-server", "foo", "bar")
		require.ErrorContains(t, err, "the server \"ze-server\" already exists")
	}

	// working with no secrets
	{
		job := SourceJob{
			Specification: &SourceJobSpecification{
				Job: &armappcontainers.Job{},
			},
		}
		err := job.SetRegistry("ze-server", "foo", "bar")
		require.NoError(t, err)
		require.Len(t, job.Specification.Job.Properties.Configuration.Secrets, 1)
		require.Len(t, job.Specification.Job.Properties.Configuration.Registries, 1)
		require.Equal(t, "azcagit-reg-cred", *job.Specification.Job.Properties.Configuration.Secrets[0].Name)
		require.Equal(t, "bar", *job.Specification.Job.Properties.Configuration.Secrets[0].Value)
		require.Equal(t, "azcagit-reg-cred", *job.Specification.Job.Properties.Configuration.Registries[0].PasswordSecretRef)
		require.Equal(t, "ze-server", *job.Specification.Job.Properties.Configuration.Registries[0].Server)
		require.Equal(t, "foo", *job.Specification.Job.Properties.Configuration.Registries[0].Username)
	}

	// working with other secrets
	{
		job := SourceJob{
			Specification: &SourceJobSpecification{
				Job: &armappcontainers.Job{
					Properties: &armappcontainers.JobProperties{
						Configuration: &armappcontainers.JobConfiguration{
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
		err := job.SetRegistry("ze-server", "foo", "bar")
		require.NoError(t, err)
		require.Len(t, job.Specification.Job.Properties.Configuration.Secrets, 1)
		require.Len(t, job.Specification.Job.Properties.Configuration.Registries, 2)
		require.Equal(t, "azcagit-reg-cred", *job.Specification.Job.Properties.Configuration.Secrets[0].Name)
		require.Equal(t, "bar", *job.Specification.Job.Properties.Configuration.Secrets[0].Value)
		require.Equal(t, "ze-other-password-ref", *job.Specification.Job.Properties.Configuration.Registries[0].PasswordSecretRef)
		require.Equal(t, "ze-other-server", *job.Specification.Job.Properties.Configuration.Registries[0].Server)
		require.Equal(t, "ze-other-username", *job.Specification.Job.Properties.Configuration.Registries[0].Username)
		require.Equal(t, "azcagit-reg-cred", *job.Specification.Job.Properties.Configuration.Registries[1].PasswordSecretRef)
		require.Equal(t, "ze-server", *job.Specification.Job.Properties.Configuration.Registries[1].Server)
		require.Equal(t, "foo", *job.Specification.Job.Properties.Configuration.Registries[1].Username)
	}

	// working with SourceJobs
	{
		job := SourceJob{
			Specification: &SourceJobSpecification{
				Job: &armappcontainers.Job{},
			},
		}
		jobs := make(SourceJobs)
		jobs["foo"] = job

		err := jobs.SetRegistry("foo", "ze-server", "foo", "bar")
		require.NoError(t, err)

		updatedJob, ok := jobs["foo"]
		require.True(t, ok)
		require.Len(t, updatedJob.Specification.Job.Properties.Configuration.Secrets, 1)
		require.Len(t, updatedJob.Specification.Job.Properties.Configuration.Registries, 1)
		require.Equal(t, "azcagit-reg-cred", *updatedJob.Specification.Job.Properties.Configuration.Secrets[0].Name)
		require.Equal(t, "bar", *updatedJob.Specification.Job.Properties.Configuration.Secrets[0].Value)
		require.Equal(t, "azcagit-reg-cred", *updatedJob.Specification.Job.Properties.Configuration.Registries[0].PasswordSecretRef)
		require.Equal(t, "ze-server", *updatedJob.Specification.Job.Properties.Configuration.Registries[0].Server)
		require.Equal(t, "foo", *updatedJob.Specification.Job.Properties.Configuration.Registries[0].Username)
	}
}

func TestSourceJobsGetRemoteSecret(t *testing.T) {
	cases := []struct {
		testDescription string
		input           *SourceJobs
		expectedOutput  []string
	}{
		{
			testDescription: "single secret",
			input: &SourceJobs{
				"foo": {
					Specification: &SourceJobSpecification{
						RemoteSecrets: []RemoteSecretSpecification{
							{
								SecretName:       toPtr("foo"),
								RemoteSecretName: toPtr("bar"),
							},
						},
					},
				},
				"bar": {
					Specification: &SourceJobSpecification{
						RemoteSecrets: []RemoteSecretSpecification{
							{
								SecretName:       toPtr("baz"),
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
			input: &SourceJobs{
				"foo": {
					Specification: &SourceJobSpecification{
						RemoteSecrets: []RemoteSecretSpecification{
							{
								SecretName:       toPtr("foo"),
								RemoteSecretName: toPtr("bar"),
							},
						},
					},
				},
				"bar": {
					Specification: &SourceJobSpecification{
						RemoteSecrets: []RemoteSecretSpecification{
							{
								SecretName:       toPtr("foo"),
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

func TestSourceJobParseLocationFilterSpecification(t *testing.T) {
	cases := []struct {
		testDescription        string
		input                  string
		expectedLocationFilter []LocationFilterSpecification
		expectedErrorContains  string
	}{
		{
			testDescription: "no locationFilter specified",
			input: `
kind: AzureContainerJob
apiVersion: aca.xenit.io/v1alpha2
metadata:
  name: foo
spec:
  job:
    properties:
      configuration:
        replicaTimeout: 1337`,
			expectedLocationFilter: nil,
			expectedErrorContains:  "",
		},
		{
			testDescription: "one locationFilter specified",
			input: `
kind: AzureContainerJob
apiVersion: aca.xenit.io/v1alpha2
metadata:
  name: foo
spec:
  locationFilter:
    - Foo Bar
  job:
    properties:
      configuration:
        replicaTimeout: 1337`,
			expectedLocationFilter: []LocationFilterSpecification{
				LocationFilterSpecification("foobar"),
			},
			expectedErrorContains: "",
		},
	}

	for i, c := range cases {
		t.Logf("Test #%d: %s", i, c.testDescription)
		job := &SourceJob{}
		isContainerJob, err := job.Unmarshal([]byte(c.input), config.Config{
			ManagedEnvironmentID: "ze-me-id",
			Location:             "zefakeregion",
		})
		require.True(t, isContainerJob)
		if c.expectedErrorContains == "" {
			require.NoError(t, err)
		} else {
			require.ErrorContains(t, err, c.expectedErrorContains)
		}

		require.Equal(t, c.expectedLocationFilter, job.Specification.LocationFilter)
	}
}

func TestSourceJobShouldRunInLocation(t *testing.T) {
	cases := []struct {
		currentLocation string
		job             *SourceJob
		expectedResult  bool
	}{
		{
			currentLocation: "foo",
			job:             nil,
			expectedResult:  true,
		},
		{
			currentLocation: "foo",
			job: &SourceJob{
				Specification: &SourceJobSpecification{
					LocationFilter: nil,
				},
			},
			expectedResult: true,
		},
		{
			currentLocation: "foo",
			job: &SourceJob{
				Specification: &SourceJobSpecification{
					LocationFilter: []LocationFilterSpecification{},
				},
			},
			expectedResult: true,
		},
		{
			currentLocation: "foo",
			job: &SourceJob{
				Specification: &SourceJobSpecification{
					LocationFilter: []LocationFilterSpecification{
						"bar",
					},
				},
			},
			expectedResult: false,
		},
		{
			currentLocation: "foo",
			job: &SourceJob{
				Specification: &SourceJobSpecification{
					LocationFilter: []LocationFilterSpecification{
						"foo",
					},
				},
			},
			expectedResult: true,
		},
		{
			currentLocation: "foo",
			job: &SourceJob{
				Specification: &SourceJobSpecification{
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
			job: &SourceJob{
				Specification: &SourceJobSpecification{
					LocationFilter: []LocationFilterSpecification{
						"foobar",
					},
				},
			},
			expectedResult: true,
		},
	}

	for _, c := range cases {
		result := c.job.ShoudRunInLocation(c.currentLocation)
		require.Equal(t, c.expectedResult, result)
	}
}

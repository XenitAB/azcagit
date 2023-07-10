package source

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
	"github.com/hashicorp/go-multierror"
	"github.com/xenitab/azcagit/src/config"
	"sigs.k8s.io/yaml"
)

const (
	AzureContainerJobVersion = "aca.xenit.io/v1alpha2"
	AzureContainerJobKind    = "AzureContainerJob"
)

type SourceJobSpecification struct {
	Job            *armappcontainers.Job         `json:"job,omitempty" yaml:"job,omitempty"`
	RemoteSecrets  []RemoteSecretSpecification   `json:"remoteSecrets,omitempty" yaml:"remoteSecrets,omitempty"`
	LocationFilter []LocationFilterSpecification `json:"locationFilter,omitempty" yaml:"locationFilter,omitempty"`
	Replacements   *ReplacementsSpecification    `json:"replacements,omitempty" yaml:"replacements,omitempty"`
}

type SourceJob struct {
	Kind          string                  `json:"kind,omitempty" yaml:"kind,omitempty"`
	APIVersion    string                  `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Metadata      map[string]string       `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Specification *SourceJobSpecification `json:"spec,omitempty" yaml:"spec,omitempty"`
	Err           error
}

func (job *SourceJob) Error() error {
	return job.Err
}

func (job *SourceJob) Name() string {
	if job.Metadata == nil {
		return ""
	}

	name, ok := job.Metadata["name"]
	if !ok {
		return ""
	}

	return name
}

func (job *SourceJob) SetSecret(name string, value string) error {
	if job == nil || job.Specification == nil || job.Specification.Job == nil {
		return fmt.Errorf("job is nil")
	}

	if job.Specification.Job.Properties == nil {
		job.Specification.Job.Properties = &armappcontainers.JobProperties{}
	}

	if job.Specification.Job.Properties.Configuration == nil {
		job.Specification.Job.Properties.Configuration = &armappcontainers.JobConfiguration{}
	}

	if job.Specification.Job.Properties.Configuration.Secrets == nil {
		job.Specification.Job.Properties.Configuration.Secrets = []*armappcontainers.Secret{}
	}

	for _, v := range job.Specification.Job.Properties.Configuration.Secrets {
		if v == nil || v.Name == nil {
			continue
		}

		if *v.Name == name {
			return fmt.Errorf("a secret with name %q already exists", name)
		}
	}

	job.Specification.Job.Properties.Configuration.Secrets = append(job.Specification.Job.Properties.Configuration.Secrets, &armappcontainers.Secret{
		Name:  &name,
		Value: &value,
	})

	return nil
}

func (job *SourceJob) SetRegistry(server string, username string, password string) error {
	if job == nil || job.Specification == nil || job.Specification.Job == nil {
		return fmt.Errorf("job is nil")
	}

	if job.Specification.Job.Properties == nil {
		job.Specification.Job.Properties = &armappcontainers.JobProperties{}
	}

	if job.Specification.Job.Properties.Configuration == nil {
		job.Specification.Job.Properties.Configuration = &armappcontainers.JobConfiguration{}
	}

	if job.Specification.Job.Properties.Configuration.Registries == nil {
		job.Specification.Job.Properties.Configuration.Registries = []*armappcontainers.RegistryCredentials{}
	}

	for _, v := range job.Specification.Job.Properties.Configuration.Registries {
		if v == nil || v.Server == nil || v.Username == nil {
			continue
		}

		if v.Identity == nil && v.PasswordSecretRef == nil {
			continue
		}

		if *v.Server == server {
			return fmt.Errorf("the server %q already exists", server)
		}
	}

	passwordSecretRef := "azcagit-reg-cred"
	err := job.SetSecret(passwordSecretRef, password)
	if err != nil {
		return err
	}

	job.Specification.Job.Properties.Configuration.Registries = append(job.Specification.Job.Properties.Configuration.Registries, &armappcontainers.RegistryCredentials{
		Server:            &server,
		PasswordSecretRef: &passwordSecretRef,
		Username:          &username,
		Identity:          nil,
	})

	return nil
}

func (job *SourceJob) GetRemoteSecrets() []RemoteSecretSpecification {
	secretsMap := make(map[string]struct{})
	if job == nil || job.Specification == nil || job.Specification.RemoteSecrets == nil || len(job.Specification.RemoteSecrets) == 0 {
		return []RemoteSecretSpecification{}
	}

	secrets := []RemoteSecretSpecification{}
	for _, secret := range job.Specification.RemoteSecrets {
		if !secret.Valid() {
			continue
		}
		secrets = append(secrets, secret)
		secretsMap[*secret.RemoteSecretName] = struct{}{}
	}

	return secrets

}

func (job *SourceJob) ValidateFields() (bool, error) {
	var result *multierror.Error
	if job.Kind == "" {
		return false, fmt.Errorf("kind is missing")
	}
	if job.Kind != "" && job.Kind != AzureContainerJobKind {
		return false, nil
	}
	requiredVersion := AzureContainerJobVersion
	if job.APIVersion != "" && job.APIVersion != requiredVersion {
		result = multierror.Append(fmt.Errorf("apiVersion for %s should be %s", job.Kind, requiredVersion), result)
	}

	if job.Specification == nil {
		result = multierror.Append(fmt.Errorf("spec is missing"), result)
	}

	if job.Specification != nil && job.Specification.Job == nil {
		result = multierror.Append(fmt.Errorf("job is missing"), result)
	}

	if job.Metadata == nil {
		result = multierror.Append(fmt.Errorf("metadata is missing"), result)
	}

	if job.Metadata != nil {
		_, ok := job.Metadata["name"]
		if !ok {
			result = multierror.Append(fmt.Errorf("name missing from metadata"), result)
		}
	}

	if job.Specification != nil && job.Specification.Job != nil && job.Specification.Job.Properties != nil && job.Specification.Job.Properties.EnvironmentID != nil {
		result = multierror.Append(fmt.Errorf("environmentID is disabled and set through azcagit"), result)
	}

	if job.Specification != nil && job.Specification.Job != nil && job.Specification.Job.Location != nil {
		result = multierror.Append(fmt.Errorf("location is disabled and set through azcagit"), result)
	}

	return true, result.ErrorOrNil()
}

func (job *SourceJob) Unmarshal(y []byte, cfg config.Config) (bool, error) {
	j, err := yaml.YAMLToJSON(y)
	if err != nil {
		return true, err
	}
	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields()
	var newjob SourceJob
	err = dec.Decode(&newjob)
	if err != nil {
		return true, err
	}

	isContainerJob, err := newjob.ValidateFields()
	if err != nil {
		return isContainerJob, err
	}

	if !isContainerJob {
		return false, nil
	}

	if cfg.ManagedEnvironmentID == "" {
		return true, fmt.Errorf("cfg.ManagedEnvironmentID is not set")
	}

	if newjob.Specification.Job.Properties == nil {
		newjob.Specification.Job.Properties = &armappcontainers.JobProperties{}
	}
	newjob.Specification.Job.Properties.EnvironmentID = &cfg.ManagedEnvironmentID

	if cfg.Location == "" {
		return true, fmt.Errorf("cfg.Location is not set")
	}
	newjob.Specification.Job.Location = &cfg.Location

	if newjob.Specification.Job.Tags == nil {
		newjob.Specification.Job.Tags = make(map[string]*string)
	}

	if len(newjob.Specification.LocationFilter) != 0 {
		sanitizedLocationFilters := []LocationFilterSpecification{}
		for _, filter := range newjob.Specification.LocationFilter {
			sanitizedLocationFilters = append(sanitizedLocationFilters, LocationFilterSpecification(sanitizeAzureLocation(filter)))
		}
		newjob.Specification.LocationFilter = sanitizedLocationFilters
	}

	newjob.Specification.Job.Tags["aca.xenit.io"] = toPtr("true")

	err = newjob.applyReplacements()
	if err != nil {
		return true, err
	}

	*job = newjob
	return true, nil
}

func (job *SourceJob) applyReplacements() error {
	if job.Specification.Replacements != nil && job.Specification.Replacements.Images != nil && len(job.Specification.Replacements.Images) != 0 {
		if job.Specification.Job.Properties.Template == nil || job.Specification.Job.Properties.Template.Containers == nil || len(job.Specification.Job.Properties.Template.Containers) == 0 {
			return fmt.Errorf("no containers found")
		}
		for i, container := range job.Specification.Job.Properties.Template.Containers {
			if container.Image == nil || *container.Image == "" {
				return fmt.Errorf("no image found for container %d", i)
			}
			oldImage := *container.Image
			imageParts := strings.Split(oldImage, ":")
			imageName := imageParts[0]
			for _, replacementImage := range job.Specification.Replacements.Images {
				if imageName == *replacementImage.ImageName {
					*job.Specification.Job.Properties.Template.Containers[i].Image = fmt.Sprintf("%s:%s", imageName, *replacementImage.NewImageTag)
				}
			}
		}
	}
	return nil
}

func (job *SourceJob) ShoudRunInLocation(currentLocation string) bool {
	if job == nil || job.Specification == nil || len(job.Specification.LocationFilter) == 0 {
		return true
	}

	fixedCurrentLocation := sanitizeAzureLocation(LocationFilterSpecification(currentLocation))
	for _, filter := range job.Specification.LocationFilter {
		if fixedCurrentLocation == filter {
			return true
		}
	}

	return false
}

type SourceJobs map[string]SourceJob

func (jobs *SourceJobs) Unmarshal(path string, y []byte, cfg config.Config) {
	if jobs == nil {
		jobs = toPtr(make(SourceJobs))
	}
	parts := strings.Split(string(y), "---")
	for i, part := range parts {
		var job SourceJob
		isContainerJob, err := job.Unmarshal([]byte(part), cfg)
		if err != nil {
			job.Err = fmt.Errorf("unable to unmarshal SourceJob from %s (document %d): %w", path, i, err)
			(*jobs)[fmt.Sprintf("%s-%d", path, i)] = job
			continue
		}
		if !isContainerJob {
			continue
		}
		_, ok := (*jobs)[job.Name()]
		if ok {
			job.Err = fmt.Errorf("unable to add %s (document %d) with name %s as name is a duplicate", path, i, job.Name())
			(*jobs)[fmt.Sprintf("%s-%d-%s", path, i, job.Name())] = job
			continue
		}
		(*jobs)[job.Name()] = job
	}
}

func (jobs *SourceJobs) GetSortedNames() []string {
	names := []string{}
	for name, job := range *jobs {
		if job.Error() != nil {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (jobs *SourceJobs) Get(name string) (SourceJob, bool) {
	job, ok := (*jobs)[name]
	if job.Error() != nil {
		return SourceJob{}, false
	}
	return job, ok
}

func (jobs *SourceJobs) Delete(name string) {
	delete(*jobs, name)
}

func (jobs *SourceJobs) SetSecret(name string, secretName string, secretValue string) error {
	job, ok := (*jobs)[name]
	if !ok {
		return fmt.Errorf("no SourceJob with name %q", name)
	}

	err := job.SetSecret(secretName, secretValue)
	if err != nil {
		return err
	}

	(*jobs)[name] = job

	return nil
}

func (jobs *SourceJobs) SetRegistry(name string, server string, username string, password string) error {
	job, ok := (*jobs)[name]
	if !ok {
		return fmt.Errorf("no SourceJob with name %q", name)
	}

	err := job.SetRegistry(server, username, password)
	if err != nil {
		return err
	}

	(*jobs)[name] = job

	return nil
}

func (jobs *SourceJobs) GetUniqueRemoteSecretNames() []string {
	secretsMap := make(map[string]struct{})
	for _, jobName := range jobs.GetSortedNames() {
		job, _ := jobs.Get(jobName)
		jobSecrets := job.GetRemoteSecrets()
		for _, remoteSecret := range jobSecrets {
			secretsMap[*remoteSecret.RemoteSecretName] = struct{}{}
		}
	}

	secrets := []string{}
	for secret := range secretsMap {
		secrets = append(secrets, secret)
	}
	sort.Strings(secrets)

	return secrets
}

func (jobs *SourceJobs) Error() error {
	var result *multierror.Error
	for _, job := range *jobs {
		if job.Error() != nil {
			result = multierror.Append(job.Error(), result)
		}
	}

	return result.ErrorOrNil()
}

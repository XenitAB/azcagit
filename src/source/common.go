package source

import (
	"strings"

	"github.com/xenitab/azcagit/src/config"
)

type RemoteSecretSpecification struct {
	SecretName       *string `json:"secretName,omitempty" yaml:"secretName,omitempty"`
	RemoteSecretName *string `json:"remoteSecretName,omitempty" yaml:"remoteSecretName,omitempty"`
}

func (r *RemoteSecretSpecification) Valid() bool {
	if r.SecretName == nil || r.RemoteSecretName == nil {
		return false
	}
	if *r.SecretName == "" || *r.RemoteSecretName == "" {
		return false
	}
	return true
}

type LocationFilterSpecification string
type ImageReplacementSpecification struct {
	ImageName   *string `json:"imageName,omitempty" yaml:"imageName,omitempty"`
	NewImageTag *string `json:"newImageTag,omitempty" yaml:"newImageTag,omitempty"`
}
type ReplacementsSpecification struct {
	Images []ImageReplacementSpecification `json:"images,omitempty" yaml:"image,omitempty"`
}

func sanitizeAzureLocation(filter LocationFilterSpecification) LocationFilterSpecification {
	filterWithoutSpaces := strings.ReplaceAll(string(filter), " ", "")
	lowercaseFilter := strings.ToLower(filterWithoutSpaces)
	return LocationFilterSpecification(lowercaseFilter)
}

func getSourcesFromFiles(yamlFiles *map[string][]byte, cfg config.Config) *Sources {
	apps := &SourceApps{}
	for path := range *yamlFiles {
		content := (*yamlFiles)[path]
		apps.Unmarshal(path, content, cfg)
	}
	jobs := &SourceJobs{}
	for path := range *yamlFiles {
		content := (*yamlFiles)[path]
		jobs.Unmarshal(path, content, cfg)
	}
	return &Sources{
		Apps: apps,
		Jobs: jobs,
	}
}

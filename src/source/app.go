package source

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
	"github.com/xenitab/aca-gitops-engine/src/config"
	"sigs.k8s.io/yaml"
)

const (
	AzureContainerAppVersion = "aca.xenit.io/v1alpha1"
	AzureContainerAppKind    = "AzureContainerApp"
)

type SourceApp struct {
	Kind          string                         `json:"kind,omitempty" yaml:"kind,omitempty"`
	APIVersion    string                         `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Metadata      map[string]string              `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Specification *armappcontainers.ContainerApp `json:"spec,omitempty" yaml:"spec,omitempty"`
}

func (app *SourceApp) Name() string {
	if app.Metadata == nil {
		return ""
	}

	name, ok := app.Metadata["name"]
	if !ok {
		return ""
	}

	return name
}

func (app *SourceApp) ValidateFields() error {
	var errs []string
	if app.Kind != "" && app.Kind != AzureContainerAppKind {
		errs = append(errs, "kind should be "+AzureContainerAppKind)
	}
	requiredVersion := AzureContainerAppVersion
	if app.APIVersion != "" && app.APIVersion != requiredVersion {
		errs = append(errs, "apiVersion for "+app.Kind+" should be "+requiredVersion)
	}

	if app.Specification == nil {
		errs = append(errs, "spec is missing")
	}

	if app.Metadata == nil {
		errs = append(errs, "metadata is missing")
	}

	if app.Metadata != nil {
		_, ok := app.Metadata["name"]
		if !ok {
			errs = append(errs, "name missing from metadata")
		}
	}

	if app.Specification != nil && app.Specification.Properties != nil && app.Specification.Properties.ManagedEnvironmentID != nil {
		errs = append(errs, "managedEnvironmentID can't be set through json")
	}

	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf(strings.Join(errs, "\n"))
}

func (app *SourceApp) Unmarshal(y []byte, cfg config.Config) error {
	j, err := yaml.YAMLToJSON(y)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields()
	var newapp SourceApp
	err = dec.Decode(&newapp)
	if err != nil {
		return err
	}

	err = newapp.ValidateFields()
	if err != nil {
		return err
	}

	if cfg.ManagedEnvironmentID != "" {
		if newapp.Specification == nil {
			newapp.Specification = &armappcontainers.ContainerApp{}
		}

		if newapp.Specification.Properties == nil {
			newapp.Specification.Properties = &armappcontainers.ContainerAppProperties{}
		}
		newapp.Specification.Properties.ManagedEnvironmentID = &cfg.ManagedEnvironmentID
	}

	if newapp.Specification.Tags == nil {
		newapp.Specification.Tags = make(map[string]*string)
	}

	newapp.Specification.Tags["aca.xenit.io"] = toPtr("true")

	*app = newapp
	return nil
}

type SourceApps map[string]SourceApp

func (apps *SourceApps) Unmarshal(y []byte, cfg config.Config) error {
	if apps == nil {
		apps = toPtr(make(SourceApps))
	}
	parts := strings.Split(string(y), "---")
	for _, part := range parts {
		var app SourceApp
		err := app.Unmarshal([]byte(part), cfg)
		if err != nil {
			return err
		}
		_, ok := (*apps)[app.Name()]
		if ok {
			return fmt.Errorf("multiple instances of %q", app.Name())
		}
		(*apps)[app.Name()] = app
	}
	return nil
}

func (apps *SourceApps) GetSortedNames() []string {
	names := []string{}
	for name := range *apps {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (apps *SourceApps) Get(name string) (SourceApp, bool) {
	app, ok := (*apps)[name]
	return app, ok
}

func getAzureContainerAppsFromFiles(yamlFiles *map[string][]byte, cfg config.Config) (*SourceApps, error) {
	apps := SourceApps{}
	for path := range *yamlFiles {
		content := (*yamlFiles)[path]
		err := apps.Unmarshal(content, cfg)
		if err != nil {
			return nil, err
		}
	}
	return &apps, nil
}

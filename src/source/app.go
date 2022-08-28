package source

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
	"github.com/hashicorp/go-multierror"
	"github.com/xenitab/azcagit/src/config"
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
	Err           error
}

func (app *SourceApp) Error() error {
	return app.Err
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

func (apps *SourceApps) Unmarshal(path string, y []byte, cfg config.Config) {
	if apps == nil {
		apps = toPtr(make(SourceApps))
	}
	parts := strings.Split(string(y), "---")
	for i, part := range parts {
		var app SourceApp
		err := app.Unmarshal([]byte(part), cfg)
		if err != nil {
			app.Err = fmt.Errorf("unable to unmarshal SourceApp from %s (document %d): %w", path, i, err)
			(*apps)[fmt.Sprintf("%s-%d", path, i)] = app
			continue
		}
		_, ok := (*apps)[app.Name()]
		if ok {
			app.Err = fmt.Errorf("unable to add %s (document %d) with name %s as name is a duplicate", path, i, app.Name())
			(*apps)[fmt.Sprintf("%s-%d-%s", path, i, app.Name())] = app
			continue
		}
		(*apps)[app.Name()] = app
	}
}

func (apps *SourceApps) GetSortedNames() []string {
	names := []string{}
	for name, app := range *apps {
		if app.Error() != nil {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (apps *SourceApps) Get(name string) (SourceApp, bool) {
	app, ok := (*apps)[name]
	if app.Error() != nil {
		return SourceApp{}, false
	}
	return app, ok
}

func (apps *SourceApps) Error() error {
	var result *multierror.Error
	for _, app := range *apps {
		if app.Error() != nil {
			result = multierror.Append(app.Error(), result)
		}
	}

	return result.ErrorOrNil()
}

func getSourceAppsFromFiles(yamlFiles *map[string][]byte, cfg config.Config) *SourceApps {
	apps := SourceApps{}
	for path := range *yamlFiles {
		content := (*yamlFiles)[path]
		apps.Unmarshal(path, content, cfg)
	}
	return &apps
}

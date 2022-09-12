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

type RemoteSecretSpecification struct {
	AppSecretName    *string `json:"appSecretName,omitempty" yaml:"appSecretName,omitempty"`
	RemoteSecretName *string `json:"remoteSecretName,omitempty" yaml:"remoteSecretName,omitempty"`
}

func (r *RemoteSecretSpecification) Valid() bool {
	if r.AppSecretName == nil || r.RemoteSecretName == nil {
		return false
	}
	if *r.AppSecretName == "" || *r.RemoteSecretName == "" {
		return false
	}
	return true
}

type LocationFilterSpecification string

type SourceAppSpecification struct {
	App            *armappcontainers.ContainerApp `json:"app,omitempty" yaml:"app,omitempty"`
	RemoteSecrets  []RemoteSecretSpecification    `json:"remoteSecrets,omitempty" yaml:"remoteSecrets,omitempty"`
	LocationFilter []LocationFilterSpecification  `json:"locationFilter,omitempty" yaml:"locationFilter,omitempty"`
}

type SourceApp struct {
	Kind          string                  `json:"kind,omitempty" yaml:"kind,omitempty"`
	APIVersion    string                  `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Metadata      map[string]string       `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Specification *SourceAppSpecification `json:"spec,omitempty" yaml:"spec,omitempty"`
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

func (app *SourceApp) SetSecret(name string, value string) error {
	if app == nil || app.Specification == nil || app.Specification.App == nil {
		return fmt.Errorf("app is nil")
	}

	if app.Specification.App.Properties == nil {
		app.Specification.App.Properties = &armappcontainers.ContainerAppProperties{}
	}

	if app.Specification.App.Properties.Configuration == nil {
		app.Specification.App.Properties.Configuration = &armappcontainers.Configuration{}
	}

	if app.Specification.App.Properties.Configuration.Secrets == nil {
		app.Specification.App.Properties.Configuration.Secrets = []*armappcontainers.Secret{}
	}

	for _, v := range app.Specification.App.Properties.Configuration.Secrets {
		if v == nil || v.Name == nil {
			continue
		}

		if *v.Name == name {
			return fmt.Errorf("a secret with name %q already exists", name)
		}
	}

	app.Specification.App.Properties.Configuration.Secrets = append(app.Specification.App.Properties.Configuration.Secrets, &armappcontainers.Secret{
		Name:  &name,
		Value: &value,
	})

	return nil
}

func (app *SourceApp) SetRegistry(server string, username string, password string) error {
	if app == nil || app.Specification == nil || app.Specification.App == nil {
		return fmt.Errorf("app is nil")
	}

	if app.Specification.App.Properties == nil {
		app.Specification.App.Properties = &armappcontainers.ContainerAppProperties{}
	}

	if app.Specification.App.Properties.Configuration == nil {
		app.Specification.App.Properties.Configuration = &armappcontainers.Configuration{}
	}

	if app.Specification.App.Properties.Configuration.Registries == nil {
		app.Specification.App.Properties.Configuration.Registries = []*armappcontainers.RegistryCredentials{}
	}

	for _, v := range app.Specification.App.Properties.Configuration.Registries {
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
	err := app.SetSecret(passwordSecretRef, password)
	if err != nil {
		return err
	}

	app.Specification.App.Properties.Configuration.Registries = append(app.Specification.App.Properties.Configuration.Registries, &armappcontainers.RegistryCredentials{
		Server:            &server,
		PasswordSecretRef: &passwordSecretRef,
		Username:          &username,
		Identity:          nil,
	})

	return nil
}

func (app *SourceApp) GetRemoteSecrets() []RemoteSecretSpecification {
	secretsMap := make(map[string]struct{})
	if app == nil || app.Specification == nil || app.Specification.RemoteSecrets == nil || len(app.Specification.RemoteSecrets) == 0 {
		return []RemoteSecretSpecification{}
	}

	secrets := []RemoteSecretSpecification{}
	for _, secret := range app.Specification.RemoteSecrets {
		if !secret.Valid() {
			continue
		}
		secrets = append(secrets, secret)
		secretsMap[*secret.RemoteSecretName] = struct{}{}
	}

	return secrets

}

func (app *SourceApp) ValidateFields() error {
	var result *multierror.Error
	if app.Kind != "" && app.Kind != AzureContainerAppKind {
		result = multierror.Append(fmt.Errorf("kind should be %s", AzureContainerAppKind), result)
	}
	requiredVersion := AzureContainerAppVersion
	if app.APIVersion != "" && app.APIVersion != requiredVersion {
		result = multierror.Append(fmt.Errorf("apiVersion for %s should be %s", app.Kind, requiredVersion), result)
	}

	if app.Specification == nil {
		result = multierror.Append(fmt.Errorf("spec is missing"), result)
	}

	if app.Specification != nil && app.Specification.App == nil {
		result = multierror.Append(fmt.Errorf("app is missing"), result)
	}

	if app.Metadata == nil {
		result = multierror.Append(fmt.Errorf("metadata is missing"), result)
	}

	if app.Metadata != nil {
		_, ok := app.Metadata["name"]
		if !ok {
			result = multierror.Append(fmt.Errorf("name missing from metadata"), result)
		}
	}

	if app.Specification != nil && app.Specification.App != nil && app.Specification.App.Properties != nil && app.Specification.App.Properties.ManagedEnvironmentID != nil {
		result = multierror.Append(fmt.Errorf("managedEnvironmentID is disabled and set through azcagit"), result)
	}

	if app.Specification != nil && app.Specification.App != nil && app.Specification.App.Location != nil {
		result = multierror.Append(fmt.Errorf("location is disabled and set through azcagit"), result)
	}

	return result.ErrorOrNil()
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

	if cfg.ManagedEnvironmentID == "" {
		return fmt.Errorf("cfg.ManagedEnvironmentID is not set")
	}

	if newapp.Specification.App.Properties == nil {
		newapp.Specification.App.Properties = &armappcontainers.ContainerAppProperties{}
	}
	newapp.Specification.App.Properties.ManagedEnvironmentID = &cfg.ManagedEnvironmentID

	if cfg.Location == "" {
		return fmt.Errorf("cfg.Location is not set")
	}
	newapp.Specification.App.Location = &cfg.Location

	if newapp.Specification.App.Tags == nil {
		newapp.Specification.App.Tags = make(map[string]*string)
	}

	if len(newapp.Specification.LocationFilter) != 0 {
		sanitizedLocationFilters := []LocationFilterSpecification{}
		for _, filter := range newapp.Specification.LocationFilter {
			sanitizedLocationFilters = append(sanitizedLocationFilters, LocationFilterSpecification(sanitizeAzureLocation(filter)))
		}
		newapp.Specification.LocationFilter = sanitizedLocationFilters
	}

	newapp.Specification.App.Tags["aca.xenit.io"] = toPtr("true")

	*app = newapp
	return nil
}

func (app *SourceApp) ShoudRunInLocation(currentLocation string) bool {
	if app == nil || app.Specification == nil || len(app.Specification.LocationFilter) == 0 {
		return true
	}

	fixedCurrentLocation := sanitizeAzureLocation(LocationFilterSpecification(currentLocation))
	for _, filter := range app.Specification.LocationFilter {
		if fixedCurrentLocation == filter {
			return true
		}
	}

	return false
}

func sanitizeAzureLocation(filter LocationFilterSpecification) LocationFilterSpecification {
	filterWithoutSpaces := strings.ReplaceAll(string(filter), " ", "")
	lowercaseFilter := strings.ToLower(filterWithoutSpaces)
	return LocationFilterSpecification(lowercaseFilter)
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

func (apps *SourceApps) Delete(name string) {
	delete(*apps, name)
}

func (apps *SourceApps) SetAppSecret(appName string, secretName string, secretValue string) error {
	app, ok := (*apps)[appName]
	if !ok {
		return fmt.Errorf("no sourceApp with name %q", appName)
	}

	err := app.SetSecret(secretName, secretValue)
	if err != nil {
		return err
	}

	(*apps)[appName] = app

	return nil
}

func (apps *SourceApps) SetAppRegistry(appName string, server string, username string, password string) error {
	app, ok := (*apps)[appName]
	if !ok {
		return fmt.Errorf("no sourceApp with name %q", appName)
	}

	err := app.SetRegistry(server, username, password)
	if err != nil {
		return err
	}

	(*apps)[appName] = app

	return nil
}

func (apps *SourceApps) GetUniqueRemoteSecretNames() []string {
	secretsMap := make(map[string]struct{})
	for _, appName := range apps.GetSortedNames() {
		app, _ := apps.Get(appName)
		appSecrets := app.GetRemoteSecrets()
		for _, remoteSecret := range appSecrets {
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

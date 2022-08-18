package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
	"sigs.k8s.io/yaml"
)

const (
	AzureContainerAppVersion = "aca.xenit.io/v1alpha1"
	AzureContainerAppKind    = "AzureContainerApp"
)

type AzureContainerApp struct {
	Kind          string                         `json:"kind,omitempty" yaml:"kind,omitempty"`
	APIVersion    string                         `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Metadata      map[string]string              `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Specification *armappcontainers.ContainerApp `json:"spec,omitempty" yaml:"spec,omitempty"`
}

func (aca *AzureContainerApp) Name() string {
	if aca.Metadata == nil {
		return ""
	}

	name, ok := aca.Metadata["name"]
	if !ok {
		return ""
	}

	return name
}

func (aca *AzureContainerApp) ValidateFields() error {
	var errs []string
	if aca.Kind != "" && aca.Kind != AzureContainerAppKind {
		errs = append(errs, "kind should be "+AzureContainerAppKind)
	}
	requiredVersion := AzureContainerAppVersion
	if aca.APIVersion != "" && aca.APIVersion != requiredVersion {
		errs = append(errs, "apiVersion for "+aca.Kind+" should be "+requiredVersion)
	}

	if aca.Specification == nil {
		errs = append(errs, "spec is missing")
	}

	if aca.Metadata == nil {
		errs = append(errs, "metadata is missing")
	}

	if aca.Metadata != nil {
		_, ok := aca.Metadata["name"]
		if !ok {
			errs = append(errs, "name missing from metadata")
		}
	}

	if aca.Specification != nil && aca.Specification.Properties != nil && aca.Specification.Properties.ManagedEnvironmentID != nil {
		errs = append(errs, "managedEnvironmentID can't be set through json")
	}

	// if aca.Specification.Properties == nil {
	// 	aca.Specification.Properties = &armappcontainers.ContainerAppProperties{
	// 		Configuration: &armappcontainers.Configuration{
	// 			Secrets:    []*armappcontainers.Secret{},
	// 			Registries: []*armappcontainers.RegistryCredentials{},
	// 			Ingress:    nil,
	// 			Dapr:       nil,
	// 		},
	// 	}
	// }

	// if aca.Specification.Properties.Configuration == nil {
	// 	aca.Specification.Properties.Configuration = &armappcontainers.Configuration{
	// 		Secrets:    []*armappcontainers.Secret{},
	// 		Registries: []*armappcontainers.RegistryCredentials{},
	// 		Ingress:    nil,
	// 		Dapr:       nil,
	// 	}
	// }

	// if aca.Specification.Properties.Configuration.Secrets == nil {
	// 	aca.Specification.Properties.Configuration.Secrets = []*armappcontainers.Secret{}
	// }
	// if aca.Specification.Properties.Configuration.Registries == nil {
	// 	aca.Specification.Properties.Configuration.Registries = []*armappcontainers.RegistryCredentials{}
	// }

	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf(strings.Join(errs, "\n"))
}

func (aca *AzureContainerApp) Unmarshal(y []byte, cfg config) error {
	j, err := yaml.YAMLToJSON(y)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields()
	var newAca AzureContainerApp
	err = dec.Decode(&newAca)
	if err != nil {
		return err
	}

	err = newAca.ValidateFields()
	if err != nil {
		return err
	}

	if cfg.ManagedEnvironmentID != "" {
		if newAca.Specification == nil {
			newAca.Specification = &armappcontainers.ContainerApp{}
		}

		if newAca.Specification.Properties == nil {
			newAca.Specification.Properties = &armappcontainers.ContainerAppProperties{}
		}
		newAca.Specification.Properties.ManagedEnvironmentID = &cfg.ManagedEnvironmentID
	}

	if newAca.Specification.Tags == nil {
		newAca.Specification.Tags = make(map[string]*string)
	}

	newAca.Specification.Tags["aca.xenit.io"] = toPtr("true")

	*aca = newAca
	return nil
}

type AzureContainerApps map[string]AzureContainerApp

func (acas *AzureContainerApps) Unmarshal(y []byte, cfg config) error {
	if acas == nil {
		acas = toPtr(make(AzureContainerApps))
	}
	parts := strings.Split(string(y), "---")
	for _, part := range parts {
		var aca AzureContainerApp
		err := aca.Unmarshal([]byte(part), cfg)
		if err != nil {
			return err
		}
		_, ok := (*acas)[aca.Name()]
		if ok {
			return fmt.Errorf("multiple instances of %q", aca.Name())
		}
		(*acas)[aca.Name()] = aca
	}
	return nil
}

func GetAzureContainerAppFromFiles(yamlFiles *YAMLFiles, cfg config) (*AzureContainerApps, error) {
	acas := AzureContainerApps{}
	for path := range *yamlFiles {
		content := (*yamlFiles)[path]
		err := acas.Unmarshal(content, cfg)
		if err != nil {
			return nil, err
		}
	}
	return &acas, nil
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
	"sigs.k8s.io/yaml"
)

const (
	AzureContainerAppVersion = "aca.xenit.io/v1alpha1"
	AzureContainerAppKind    = "AzureContainerApp"
)

type AzureContainerApp struct {
	Kind         string                         `json:"kind,omitempty" yaml:"kind,omitempty"`
	APIVersion   string                         `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Name         string                         `json:"name,omitempty" yaml:"name,omitempty"`
	ContainerApp *armappcontainers.ContainerApp `json:"containerapp,omitempty" yaml:"containerapp,omitempty"`
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

	if aca.ContainerApp != nil && aca.ContainerApp.Properties != nil && aca.ContainerApp.Properties.ManagedEnvironmentID != nil {
		errs = append(errs, "managedEnvironmentID can't be set through json")
	}

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
		if newAca.ContainerApp == nil {
			newAca.ContainerApp = &armappcontainers.ContainerApp{}
		}
		if newAca.ContainerApp.Properties == nil {
			newAca.ContainerApp.Properties = &armappcontainers.ContainerAppProperties{}
		}
		newAca.ContainerApp.Properties.ManagedEnvironmentID = &cfg.ManagedEnvironmentID
	}

	if newAca.ContainerApp.Tags == nil {
		newAca.ContainerApp.Tags = make(map[string]*string)
	}

	newAca.ContainerApp.Tags["aca.xenit.io"] = toPtr("true")

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
		_, ok := (*acas)[aca.Name]
		if ok {
			return fmt.Errorf("multiple instances of %q", aca.Name)
		}
		(*acas)[aca.Name] = aca
	}
	return nil
}

func GetAzureContainerAppFromFiles(files []string, cfg config) (*AzureContainerApps, error) {
	acas := AzureContainerApps{}
	for _, file := range files {
		path := fmt.Sprintf("%s/%s", cfg.YAMLPath, file)
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		err = acas.Unmarshal(b, cfg)
		if err != nil {
			return nil, err
		}
	}
	return &acas, nil
}

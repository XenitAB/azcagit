//go:build generateschema

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"
	"github.com/invopop/yaml"
	"github.com/xenitab/azcagit/src/source"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "schema generation returned an error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	err := generateSchema(&source.SourceApp{}, "app")
	if err != nil {
		return err
	}

	err = generateSchema(&source.SourceJob{}, "job")
	if err != nil {
		return err
	}

	return nil
}

func generateSchema[T any](t T, name string) error {
	schema := jsonschema.Reflect(&t)
	schemaJson, err := schema.MarshalJSON()
	if err != nil {
		return err
	}

	var schemaJsonRaw interface{}
	err = json.Unmarshal(schemaJson, &schemaJsonRaw)
	if err != nil {
		return err
	}

	prettySchemaJson, err := json.MarshalIndent(schemaJsonRaw, "", "  ")
	if err != nil {
		return err
	}

	schemaYaml, err := yaml.JSONToYAML(schemaJson)
	if err != nil {
		return err
	}

	currDir, err := os.Getwd()
	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s/schemas/%s.json", currDir, name), prettySchemaJson, 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s/schemas/%s.yaml", currDir, name), schemaYaml, 0644)
	if err != nil {
		return err
	}

	return nil
}

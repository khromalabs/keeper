package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"

	// "khromalabs/keeper/storage"
	// "khromalabs/keeper/storage/sqlite"
	// "keeperUI"
)

// Conf is ...
type Conf struct {
	Paths map[string]string
}

var conf = Conf{
	Paths: map[string]string{
		"templates": "templates/",
	},
}

// go:embed templateSchema.json
var jsonTemplateSchemaData string

func main() {
	args := os.Args[1:]

	if len(args) != 1 {
		println("Usage: keeper <template>")
		return
	}

	template, err := parseTemplate(conf.Paths["templates"] + args[0] + ".yaml")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%+v\n", template)
	// keeperUI.add(template)
}

func parseTemplate(templateFilename string) (map[string]interface{}, error) {
	var yamlTemplateData map[string]interface{}

	yamlTemplateFile, err := ioutil.ReadFile(templateFilename)
	if err != nil {
		return nil, fmt.Errorf("Failed to read template YAML file: %v", err)
	}
	err = yaml.Unmarshal(yamlTemplateFile, &yamlTemplateData)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse template YAML file: %v", err)
	}
	jsonTemplateData, err := json.Marshal(yamlTemplateData)
	if err != nil {
		return nil, fmt.Errorf("Error converting template YAML to JSON: %v", err)
	}
	result, err := gojsonschema.Validate(
		gojsonschema.NewStringLoader(jsonTemplateSchemaData),
		gojsonschema.NewStringLoader(string(jsonTemplateData)),
	)
	if err != nil {
		return nil, fmt.Errorf("Error validating JSON converted YAML template against JSON Schema: %v", err)
	}
	if !result.Valid() {
		errmsg := "YAML template is not valid. See errors:\n"
		for _, desc := range result.Errors() {
			errmsg += fmt.Sprintf("- %s\n", desc)
		}
		return nil, fmt.Errorf(errmsg)
	}
	return yamlTemplateData, nil
}

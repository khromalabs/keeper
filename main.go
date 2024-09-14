package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"khromalabs/keeper/internal/config"
	. "khromalabs/keeper/internal/log"
	"khromalabs/keeper/internal/storage"
	"khromalabs/keeper/internal/ui"
	"os"
	"regexp"
	"strings"

	"github.com/pborman/getopt/v2"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
	// "github.com/davecgh/go-spew/spew"
)

//go:embed resources/schemas/templates.json
var templateJsonSchema string

//go:embed resources/texts/appInfo.txt
var appInfo string

var conf *config.Config

var version string

func main() {
	arg, err := parseArguments()
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	// LogD.Printf("arg: %#v\n", arg)
	if err := config.Check(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %s\n", err)
		os.Exit(1)
	}
	conf = config.Get()
	if arg[0] == "templates" {
		if len(arg[1]) > 0 {
			listTemplateFields(arg[1])
		} else {
			listTemplates()
		}
	} else if err := processCmd(arg[0], arg[1], arg[2]); err != nil {
		fmt.Fprintln(os.Stderr, "Error processing command:", err)
		os.Exit(1)
	}
}

func parseArguments() ([]string, error) {
	appInfoHeader := "Usage:\tkeeper <template> [-r|--read|-R|--READ|-u|--update|-d|--delete] [filter exp]\n\tkeeper -t [template]\n\tkeeper -h|--help\n\tkeeper -v|--version"
	if len(os.Args) < 2 {
		return nil, fmt.Errorf(appInfoHeader)
	}
	if os.Args[1] == "-t" {
		var arg2 string
		if len(os.Args) > 2 {
			arg2 = os.Args[2]
		} else {
			arg2 = ""
		}
		return []string{"templates", arg2, ""}, nil
	} else if os.Args[1] == "-h" || os.Args[1] == "--help" {
		return nil, fmt.Errorf(appInfo)
	} else if os.Args[1] == "-v" || os.Args[1] == "--version" {
		return nil, fmt.Errorf("Keeper version %s\n", version)
	}
	var regex = "^[A-z][A-z0-9_-]+$"
	if match, err := regexp.MatchString(regex, os.Args[1]); err != nil || !match {
		return nil, fmt.Errorf(appInfoHeader + "Invalid template (Expected: word)")
	}
	template := os.Args[1]
	os.Args = append(os.Args[:1], os.Args[2:]...)
	var options [4]*bool
	options[0] = getopt.BoolLong("read", 'r', "Read registries")
	options[1] = getopt.BoolLong("update", 'u', "Update registries")
	options[2] = getopt.BoolLong("delete", 'd', "Delete registries")
	options[3] = getopt.BoolLong("READ", 'R', "Read registries (ignoring long fields)")
	if err := getopt.Getopt(nil); err != nil {
		return nil, fmt.Errorf(appInfoHeader + "Error parsing optional arguments: " + fmt.Sprintf("%s\n", err))
	}
	n := 0
	for _, o := range options {
		if *o {
			n++
		}
	}
	if n > 1 {
		return nil, fmt.Errorf(appInfoHeader + "Please specify only one operation (read, update or delete)")
	}
	operation := "create"
	if *options[0] {
		operation = "read"
	} else if *options[1] {
		operation = "update"
	} else if *options[2] {
		operation = "delete"
	} else if *options[3] {
		operation = "READ"
	}
	expression := strings.Join(getopt.Args(), " ")
	if expression != "" && operation == "create" {
		operation = "read"
	}
	return []string{template, operation, expression}, nil
}

func processCmd(template string, action string, filter string) error {
	templateSchema, templateKeys, err := loadTemplate(
		conf.Path["templates"] + template + ".yaml")
	if err != nil {
		return fmt.Errorf("Error loading template: %s\n", err)
	}
	// LogD.Printf("templateSchema: %+v\n", templateSchema)
	// LogD.Printf("templateKeys: %+v\n", templateKeys)
	storage, err := storage.Init(template, templateSchema)
	if err != nil {
		return fmt.Errorf("Error initializing storage: %s\n", err)
	}
	// LogD.Printf("storage: %+v\n", storage)
	ui, err := ui.Init(template, templateSchema, templateKeys)
	if err != nil {
		return fmt.Errorf("Error initializing UI: %s\n", err)
	}
	// LogD.Printf("ui: %+v\n", ui)
	isUpdateOp := false
	switch action {
	case "create":
		if err := ui.Create(storage.Create); err != nil {
			return err
		}
	case "update":
		isUpdateOp = true
		fallthrough
	case "delete":
		err := ui.UpdateOrDelete(filter, isUpdateOp, storage.Read, storage.UpdateOrDelete)
		if err != nil {
			return err
		}
	default: // read
		if action == "READ" {
			conf.Miniread = true
			LogD.Printf("miniread: %#v", conf.Miniread)
		}
		if err := ui.Print(filter, storage.Read); err != nil {
			return err
		}
	}
	fmt.Println()
	return nil
}

type yamlKeysSlice struct {
	keys []string
}

func (m *yamlKeysSlice) UnmarshalYAML(value *yaml.Node) error {
	m.keys = make([]string, 0)
	for i := 0; i < len(value.Content); i++ {
		t := strings.Trim(value.Content[i].Value, " ")
		if t != "" {
			m.keys = append(m.keys, value.Content[i].Value)
		}
	}
	return nil
}

func loadTemplate(fileName string) (map[string]interface{}, []string, error) {
	file, err := os.ReadFile(fileName)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to read template YAML file: %v", err)
	}
	var yamlKeys yamlKeysSlice
	if err = yaml.Unmarshal(file, &yamlKeys); err != nil {
		return nil, nil, fmt.Errorf("Failed to parse template YAML file: %v", err)
	}
	var yamlData map[string]interface{}
	if err = yaml.Unmarshal(file, &yamlData); err != nil {
		return nil, nil, fmt.Errorf("Failed to parse template YAML file: %v", err)
	}
	jsonData, err := json.Marshal(yamlData)
	if err != nil {
		return nil, nil, fmt.Errorf("Error converting template YAML to JSON: %v", err)
	}
	result, err := gojsonschema.Validate(
		gojsonschema.NewStringLoader(templateJsonSchema),
		gojsonschema.NewStringLoader(string(jsonData)),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("Error validating YAML storage template: %v", err)
	}
	if !result.Valid() {
		errmsg := "YAML template is not valid. See errors:\n"
		for _, desc := range result.Errors() {
			errmsg += fmt.Sprintf("- %s\n", desc)
		}
		return nil, nil, fmt.Errorf(errmsg)
	}
	return yamlData, yamlKeys.keys, nil
}

func listTemplates() error {
	entries, err := os.ReadDir(conf.Path["templates"] + "./")
	if err != nil {
		return fmt.Errorf("Can't read templates!")
	}
	for _, e := range entries {
		s := strings.Split(e.Name(), ".")
		fmt.Printf("%s\n", s[0])
	}
	return nil
}

func listTemplateFields(template string) error {
	_, templatekeys, err := loadTemplate(
		conf.Path["templates"] + template + ".yaml")
	if err != nil {
		return fmt.Errorf("Error loading template: %s\n", err)
	}
	fmt.Printf("id:")
	for _, key := range templatekeys {
		fmt.Printf("\n%s:", key)
	}
	fmt.Println()
	return nil
}

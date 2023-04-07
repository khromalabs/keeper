package main


import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"github.com/khromalabs/keeper/storage"
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

func parseTemplate(filename string) (map[string]interface{}, error) {
	var data map[string]interface{}

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %v", err)
	}

	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML file: %v", err)
	}

	return data, nil
}

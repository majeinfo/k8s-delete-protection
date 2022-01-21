package main

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// Rules are "must" or "must_not" but have the same syntax
// TODO: does not specify the apiversion/group
type Rule struct {
	Namespace string `yaml:"namespace"`
	Kinds []string `yaml:"kinds"`
	Label string	`yaml:"label"`
}

func load_rules_file(filename string) []Rule {
	var rules []Rule
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Could not read file %s: %v", filename, err)
	}

	err = yaml.UnmarshalStrict([]byte(data), &rules)
	if err != nil {
		log.Fatalf("YAML error: %v", err)
	}
	log.Debugf("Rules: %v", rules)

	return rules
}

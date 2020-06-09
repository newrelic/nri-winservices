package nri

import (
	"io/ioutil"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-winservices/src/matcher"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Matcher matcher.Matcher
}

type ConfigYml struct {
	FilterEntity map[string][]string `yaml:"filter_entity"`
}

func ParseConfigYaml(filename string) (*Config, error) {
	// Read the file
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Error("Failed to open %s: %s", filename, err)
		return nil, err
	}
	// Parse the file
	c := ConfigYml{FilterEntity: make(map[string][]string)}
	if err := yaml.Unmarshal(yamlFile, &c); err != nil {
		log.Error("Failed to parse config: %s", err)
		return nil, err
	}
	log.Debug("filter :%v", c.FilterEntity["windowsService.name"])

	m := matcher.New(c.FilterEntity["windowsService.name"])

	return &Config{Matcher: m}, nil
}

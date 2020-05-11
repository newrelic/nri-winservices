package nri

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Rules struct {
	EntityRules []EntityRules `yaml:"entities"`
}
type EntityRules struct {
	EntityType string        `yaml:"type"`
	Name       string        `yaml:"name"`
	Metrics    []MetricRules `yaml:"metrics"`
}
type MetricRules struct {
	ProviderName string   `yaml:"provider_name"`
	MetricType   string   `yaml:"type"`
	NrdbName     string   `yaml:"nrdb_name"`
	SkipValue    float64  `yaml:"skip_value"`
	Attributes   []string `yaml:"attributes"`
}

func loadRules(file string) *Rules {
	rulesFile, err := ioutil.ReadFile(file)
	fatalOnErr(err)

	var rules Rules
	err = yaml.Unmarshal(rulesFile, &rules)
	fatalOnErr(err)

	return &rules
}

func (r *EntityRules) getMetricRules(providerName string) (*MetricRules, error) {
	for _, m := range r.Metrics {
		if m.ProviderName == providerName {
			return &m, nil //todo check if copy
		}
	}
	return nil, fmt.Errorf("No rules find for providerName: %s", providerName)
}

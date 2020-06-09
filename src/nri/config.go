package nri

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-winservices/src/matcher"
	yaml "gopkg.in/yaml.v2"
)

const (
	minScrapeInterval = 15 * time.Second
	heartBeatPeriod   = 5 * time.Second // Period for the hard beat signal should be less than timeout
)

// Config holds the integration configuration
type Config struct {
	Matcher             matcher.Matcher
	ExporterBindAddress string
	ExporterBindPort    string
	ScrapeInterval      time.Duration
	HeartBeatPeriod     time.Duration
}

type configYml struct {
	FilterEntity        map[string][]string `yaml:"filter_entity"`
	ExporterBindAddress string              `yaml:"exporter_bind_address"`
	ExporterBindPort    string              `yaml:"exporter_bind_port"`
	ScrapeInterval      string              `yaml:"scrape_interval"`
}

// NewConfig reads the configuration from yml file
func NewConfig(filename string) (*Config, error) {
	// Read the file
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %s", filename, err)
	}
	// Parse the file
	c := configYml{FilterEntity: make(map[string][]string)}
	if err := yaml.Unmarshal(yamlFile, &c); err != nil {
		return nil, fmt.Errorf("failed to parse config: %s", err)
	}
	var m matcher.Matcher
	if val, ok := c.FilterEntity["windowsService.name"]; ok {
		m = matcher.New(val)
	} else {
		return nil, fmt.Errorf("failed to parse config: only filter by windowsService.name is allowed")
	}
	if m.IsEmpty() {
		return nil, fmt.Errorf("failed to parse config: no valid filter loaded")
	}

	if c.ExporterBindAddress == "" || c.ExporterBindPort == "" {
		return nil, fmt.Errorf("exporter_bind_address and exporter_bind_port need to be configured")
	}

	interval, err := time.ParseDuration(c.ScrapeInterval)
	if err != nil {
		log.Error("error parsing scrape interval:%s", err.Error())
		interval = minScrapeInterval
	}
	if interval < minScrapeInterval {
		log.Warn("scrap interval defined is less than 15s. Interval has set to 15s ")
		interval = minScrapeInterval
	}
	log.Debug("running with scrape interval: %s", interval.String())

	config := &Config{
		Matcher:             m,
		ExporterBindAddress: c.ExporterBindAddress,
		ExporterBindPort:    c.ExporterBindPort,
		ScrapeInterval:      interval,
		HeartBeatPeriod:     heartBeatPeriod,
	}
	return config, nil
}

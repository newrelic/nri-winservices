/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package nri

import (
	"fmt"
	"os"
	"time"

	"github.com/newrelic/infra-integrations-sdk/v4/log"
	"github.com/newrelic/nri-winservices/src/matcher"
	yaml "gopkg.in/yaml.v2"
)

const (
	minScrapeInterval = 15 * time.Second
	heartBeatPeriod   = 5 * time.Second // Period for the hard beat signal should be less than timeout

	filterServiceName = "windowsService.name"
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
	FilterEntity        map[string][]string `yaml:"include_matching_entities"`
	ExporterBindAddress string              `yaml:"exporter_bind_address"`
	ExporterBindPort    string              `yaml:"exporter_bind_port"`
	ScrapeInterval      string              `yaml:"scrape_interval"`
}

// NewConfig reads the configuration from yml file
func NewConfig(filename string) (*Config, error) {
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", filename, err)
	}
	// Parse the file
	c := configYml{FilterEntity: make(map[string][]string)}
	if err := yaml.Unmarshal(yamlFile, &c); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	val, exists := c.FilterEntity[filterServiceName]
	if !exists {
		return nil, fmt.Errorf("%s filter inside 'include_matching_entities' config is required", filterServiceName)

	}

	matcher := matcher.New(val)
	if matcher.IsEmpty() {
		return nil, fmt.Errorf("include_matching_entities must contain at least one valid filter")
	}

	if c.ExporterBindAddress == "" || c.ExporterBindPort == "" {
		return nil, fmt.Errorf("exporter_bind_address and exporter_bind_port config must exist")
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
		Matcher:             matcher,
		ExporterBindAddress: c.ExporterBindAddress,
		ExporterBindPort:    c.ExporterBindPort,
		ScrapeInterval:      interval,
		HeartBeatPeriod:     heartBeatPeriod,
	}
	return config, nil
}

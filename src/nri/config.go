//go:build windows && amd64
// +build windows,amd64

/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package nri

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/newrelic/infra-integrations-sdk/v4/log"
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
	IncludeEntity       map[string][]string `yaml:"include_matching_entities"`
	ExcludeEntity       map[string][]string `yaml:"exclude_matching_entities"`
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
	c := configYml{
		IncludeEntity: make(map[string][]string),
		ExcludeEntity: make(map[string][]string),
	}
	if err := yaml.Unmarshal(yamlFile, &c); err != nil {
		return nil, fmt.Errorf("failed to parse config: %s", err)
	}

	var m matcher.Matcher
	var includeFilters, excludeFilters []string

	// Get include filters
	if val, ok := c.IncludeEntity["windowsService.name"]; ok {
		includeFilters = val
	}

	// Get exclude filters
	if val, ok := c.ExcludeEntity["windowsService.name"]; ok {
		excludeFilters = val
	}

	// Must have at least include filters (exclude-only is not supported)
	if len(includeFilters) == 0 {
		return nil, fmt.Errorf("failed to parse config: include_matching_entities is required for windowsService.name (exclude-only filtering is not supported)")
	}

	// Create matcher with both include and exclude filters
	m = matcher.NewWithIncludesExcludes(includeFilters, excludeFilters)
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

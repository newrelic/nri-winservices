/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/newrelic/nri-winservices/src/exporter"
	"github.com/newrelic/nri-winservices/src/nri"
	"github.com/newrelic/nri-winservices/src/scraper"

	"github.com/newrelic/infra-integrations-sdk/v4/integration"
	"github.com/newrelic/infra-integrations-sdk/v4/log"

	//This import is useful merely to keep track of dependency and generate license automatically
	_ "github.com/newrelic-forks/windows_exporter/collector"
)

type argumentList struct {
	Version    bool   `default:"false" help:"Print the integration version and commit hash"`
	Verbose    bool   `default:"false" help:"Print more information to logs."`
	Pretty     bool   `default:"false" help:"Print pretty formatted JSON."`
	ConfigPath string `default:"" help:"Path to the config file."`
}

const (
	integrationName = "com.newrelic.winservices"
)

var (
	args               argumentList
	integrationVersion = "v0.0.0"  // set by -ldflags on build
	commitHash         = "default" // Commit hash used to build the integration set by -ldflags on build
)

func main() {
	i, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	fatalOnErr(err)
	log.SetupLogging(args.Verbose)

	v := fmt.Sprintf("integration version: %s commit: %s", integrationVersion, commitHash)
	if args.Version {
		fmt.Print(v)
		return
	}
	log.Debug(v)

	config, err := nri.NewConfig(args.ConfigPath)
	fatalOnErr(err)

	e, err := exporter.New(args.Verbose, config.ExporterBindAddress, config.ExporterBindPort)
	fatalOnErr(err)

	log.Debug("Running exporter")
	err = e.Run()
	fatalOnErr(err)

	// After fail the integration is being relaunched by the Agent when timeout expires since no heartbeats are send
	log.Debug("Running Integration")
	err = run(e, i, config)
	log.Fatal(err)
}

func run(e *exporter.Exporter, i *integration.Integration, config *nri.Config) error {
	defer e.Kill()
	heartBeat := time.NewTicker(config.HeartBeatPeriod)
	metricInterval := time.NewTicker(config.ScrapeInterval)

	for {
		select {
		case <-heartBeat.C:
			log.Debug("Sending heartBeat")
			// hart beat signal for long running integrations
			// https://docs.newrelic.com/docs/integrations/integrations-sdk/file-specifications/host-integrations-newer-configuration-format#timeout
			fmt.Println("{}")

		case <-metricInterval.C:
			t := time.Now()
			log.Debug("Scraping and publishing metrics")

			metricsByFamily, err := scraper.Get(http.DefaultClient, "http://"+e.URL+e.MetricPath)
			if err != nil {
				return fmt.Errorf("fail to scrape metrics:%v", err)
			}
			log.Debug("Metrics scraped, MetricsByFamily found: %d, time elapsed: %s", len(metricsByFamily), time.Since(t).String())

			if err = nri.ProcessMetrics(i, metricsByFamily, config.Matcher); err != nil {
				return fmt.Errorf("fail to process metrics:%v", err)
			}
			log.Debug("Metrics processed, entities found: %d, time elapsed: %s", len(i.Entities), time.Since(t).String())

			err = i.Publish()
			if err != nil {
				log.Error("failed to publish integration:%v", err)
			}
			log.Debug("Metrics published")

		case <-e.Done:
			log.Debug("The exporter is not running anymore, the integration is going to be stopped")
			// exit when the exporter has stopped running
			return fmt.Errorf("exporter has stopped")
		}
	}
}

func fatalOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

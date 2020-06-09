package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/newrelic/nri-winservices/src/exporter"
	"github.com/newrelic/nri-winservices/src/nri"
	"github.com/newrelic/nri-winservices/src/scraper"

	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
)

type argumentList struct {
	Version    bool `default:"false" help:"Print the integration version and commit hash"`
	Verbose    bool `default:"false" help:"Print more information to logs."`
	Pretty     bool `default:"false" help:"Print pretty formatted JSON."`
	ConfigPath string
	// FilterList          string `default:"" help:"List of filter that are used to filter the services that are sent by the integration."`
	ExporterBindAddress string `default:"" help:"The IP address to bind to for the Prometheus exporter launched by this integration."`
	ExporterBindPort    string `default:"" help:"Binding port of the Prometheus exporter launched by this integration."`
	ScrapeInterval      string `default:"30s" help:"Interval of time for scraping metrics from the prometheus exporter. es: 30s"`
}

const (
	integrationName   = "com.newrelic.winservices"
	heartBeatPeriod   = 5 * time.Second // Period for the hard beat signal should be less than timeout
	minScrapeInterval = 15 * time.Second
)

var (
	args argumentList
	// args               sdkArgs.DefaultArgumentList
	integrationVersion = "0.0.0"   // set by -ldflags on build
	commitHash         = "default" // Commit hash used to build the integration set by -ldflags on build
)

func main() {
	i, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	if err != nil {
		log.Fatal(err)
	}
	log.SetupLogging(args.Verbose)

	v := fmt.Sprintf("integration version: %s commit: %s", integrationVersion, commitHash)
	if args.Version {
		fmt.Print(v)
		return
	}
	log.Debug(v)

	if args.ExporterBindAddress == "" || args.ExporterBindPort == "" {
		log.Fatal(fmt.Errorf("exporter_bind_address and exporter_bind_port need to be configured"))
	}

	interval, err := time.ParseDuration(args.ScrapeInterval)
	if err != nil {
		log.Error(err.Error())
		interval = minScrapeInterval
	}
	if interval < minScrapeInterval {
		log.Warn("scrap interval defined is less than 15s. Interval has set to 15s ")
		interval = minScrapeInterval
	}
	log.Debug("Running with scrape interval: %s", interval.String())

	e, err := exporter.New(args.Verbose, args.ExporterBindAddress, args.ExporterBindPort)
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Running exporter")
	err = e.Run()
	if err != nil {
		log.Fatal(err)
	}

	// After fail the integration is being relaunched by the Agent when timeout expires since no heartbeats are send
	log.Debug("Running Integration")
	err = run(e, i, interval)
	log.Fatal(err)
}

func run(e *exporter.Exporter, i *integration.Integration, interval time.Duration) error {
	defer e.Kill()
	heartBeat := time.NewTicker(heartBeatPeriod)
	metricInterval := time.NewTicker(interval)
	// log.Debug("Filter list: %s", args.FilterList)
	// matcher := matcher.New(args.FilterList)
	c, _ := nri.ParseConfigYaml(args.ConfigPath)
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

			if err = nri.ProcessMetrics(i, metricsByFamily, c.Matcher); err != nil {
				return fmt.Errorf("fail to process metrics:%v", err)
			}
			log.Debug("Metrics processed, entities found: %d, time elapsed: %s", len(i.Entities), time.Since(t).String())

			if err = nri.ProcessInventory(i); err != nil {
				return fmt.Errorf("fail to process inventory:%v", err)
			}
			log.Debug("Inventory processed, time elapsed: %s", time.Since(t).String())

			err = i.Publish()
			if err != nil {
				log.Error("failed to publish integration:%v", err)
			}
			log.Debug("Metrics and inventory published")

		case <-e.Done:
			log.Debug("The exporter is not running anymore, the integration is going to be stopped")
			// exit when the exporter has stopped running
			return fmt.Errorf("exporter has stopped")
		}
	}
}

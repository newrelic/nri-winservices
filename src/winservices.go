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
	Version             bool   `default:"false" help:"Print the integration version and commit hash"`
	Verbose             bool   `default:"false" help:"Print more information to logs."`
	Pretty              bool   `default:"false" help:"Print pretty formatted JSON."`
	AllowList           string `default:"" help:"Comma separated list of names of services to be included. By default no service is included"`
	AllowRegex          string `default:"" help:"If set, the Regex specified will be applied to filter in services. es : \"^win\" will include all services starting with \"win\"."`
	DenyList            string `default:"" help:"Comma separated list of names of services to be excluded. This is the last rule applied that take precedence over -allowList and -allowRegex"`
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
	args               argumentList
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
	}
	if interval < minScrapeInterval {
		log.Warn("scrap interval defined is less than 15s. Interval has set to 15s ")
		interval = minScrapeInterval
	}
	e := exporter.New(args.Verbose, args.ExporterBindAddress, args.ExporterBindPort)
	if err = e.Run(); err != nil {
		log.Fatal(err)
	}
	// After fail the integration is being relaunched by the Agent when timeout expires since no hartbeats are send
	log.Fatal(run(e, i, interval))
}

func run(e exporter.Exporter, i *integration.Integration, interval time.Duration) error {
	defer e.Kill()
	heartBeat := time.NewTicker(heartBeatPeriod)
	metricInterval := time.NewTicker(interval)
	for {
		select {
		case <-heartBeat.C:
			// hart beat signal for long running integrations
			// https://docs.newrelic.com/docs/integrations/integrations-sdk/file-specifications/host-integrations-newer-configuration-format#timeout
			fmt.Println("{}")

		case <-metricInterval.C:
			metricsByFamily, err := scraper.Get(http.DefaultClient, "http://"+e.URL+e.MetricPath)
			if err != nil {
				return fmt.Errorf("fail to scrape metrics:%v", err)
			}
			validator := nri.NewValidator(args.AllowList, args.DenyList, args.AllowRegex)
			if err = nri.ProcessMetrics(i, metricsByFamily, validator); err != nil {
				return fmt.Errorf("fail to process metrics:%v", err)
			}
			if err = nri.ProcessInventory(i); err != nil {
				return fmt.Errorf("fail to process inventory:%v", err)
			}
			err = i.Publish()
			if err != nil {
				log.Error("failed to publish integration:%v", err)
			}

		case <-e.Done:
			// exit when the exporter has stopped running
			return fmt.Errorf("exporter has stopped")
		}
	}
}

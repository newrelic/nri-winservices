package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/newrelic/nri-winservices/src/exporter"
	"github.com/newrelic/nri-winservices/src/nri"
	"github.com/newrelic/nri-winservices/src/scraper"

	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
)

type argumentList struct {
	Verbose             bool   `default:"false" help:"Print more information to logs."`
	Pretty              bool   `default:"false" help:"Print pretty formatted JSON."`
	AllowList           string `default:"" help:"Comma separated list of names of services to be included. By default no service is included"`
	AllowRegex          string `default:"" help:"If set, the Regex specified will be applied to filter in services. es : \"^win\" will include all services starting with \"win\"."`
	DenyList            string `default:"" help:"Comma separated list of names of services to be excluded. This is the last rule applied that take precedence over -allowList and -allowRegex"`
	ExporterBindAddress string `default:"127.0.0.1" help:"The IP address to bind to for the Prometheus exporter launched by this integration. Default is 127.0.0.1"`
	ExporterBindPort    string `default:"9182" help:"Binding port of the Prometheus exporter launched by this integration. Default is 9182"`
	ScrapeInterval      string `default:"30s" help:"Interval of time for scraping metrics from the prometheus exporter. es: 30s"`
}

const (
	integrationName    = "com.newrelic.winservices"
	integrationVersion = "0.0.2"
	heartBeatPeriod    = time.Second // Period for the hard beat signal should be less than timeout
)

var (
	args argumentList
)

func main() {
	log.SetupLogging(args.Verbose)

	i, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	e := exporter.New(args.Verbose, args.ExporterBindAddress, args.ExporterBindPort)
	if err = e.Run(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	run(e, i)
	// Integration exit if there are problems scraping or processing metrics and
	// is being relaunched by the Agent since no hartbeats are send
	os.Exit(1)
}

func run(e exporter.Exporter, i *integration.Integration) {
	interval, err := time.ParseDuration(args.ScrapeInterval)
	if err != nil {
		log.Error("error parsing ScrapeInterval:", err.Error())
		os.Exit(1)
	}
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
				log.Error("fail to scrape metrics:%v", err.Error())
				return
			}
			validator := nri.NewValidator(args.AllowList, args.DenyList, args.AllowRegex)
			if err = nri.ProcessMetrics(i, metricsByFamily, validator); err != nil {
				log.Error("fail to process metrics:%v", err.Error())
				return
			}
			if err = nri.ProcessInventory(i); err != nil {
				log.Error("fail to process inventory:%v", err.Error())
				return
			}
			err = i.Publish()
			logOnErr(err)

		case <-e.Done:
			// exit when the exporter has stopped running
			log.Error("exporter has stopped")
			return
		}
	}
}

func logOnErr(err error) {
	if err != nil {
		log.Error(err.Error())
	}
}

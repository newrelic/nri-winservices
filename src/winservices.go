package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
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
		logOnErr(err)
		os.Exit(1)
	}

	e := exporter.New(args.Verbose, args.ExporterBindAddress, args.ExporterBindPort)
	e.Run()

	run(e, i)
}

func run(e exporter.Exporter, i *integration.Integration) {
	interval, err := time.ParseDuration(args.ScrapeInterval)
	logOnErr(err)
	heartBeat := time.NewTicker(heartBeatPeriod)
	metricInterval := time.NewTicker(interval)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for {
		select {
		case <-heartBeat.C:
			// hart beat signal for long running integrations
			// https://docs.newrelic.com/docs/integrations/integrations-sdk/file-specifications/host-integrations-newer-configuration-format#timeout
			fmt.Println("{}")

		case <-metricInterval.C:
			metricsByFamily, err := scraper.Get(http.DefaultClient, "http://"+e.ExporterURL+"/metrics")
			logOnErr(err)
			validator := nri.NewValidator(args.AllowList, args.DenyList, args.AllowRegex)
			err = nri.ProcessMetrics(i, metricsByFamily, validator)
			logOnErr(err)
			err = nri.ProcessInventory(i)
			logOnErr(err)
			err = i.Publish()
			logOnErr(err)

		case osCall := <-c:
			log.Info("gracefully shuting down on system call:%+v", osCall)
			e.Kill()
			return
		}
	}
}

func logOnErr(err error) {
	if err != nil {
		log.Error(err.Error())
	}
}

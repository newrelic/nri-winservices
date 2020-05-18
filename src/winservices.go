package main

import (
	"net/http"
	"os"

	"github.com/newrelic/nri-winservices/src/nri"
	"github.com/newrelic/nri-winservices/src/scraper"

	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
)

type argumentList struct {
	Verbose     bool   `default:"false" help:"Print more information to logs."`
	Pretty      bool   `default:"false" help:"Print pretty formatted JSON."`
	ExporterURL string `default:"http://localhost:9182/metrics" help:"The url to which the scraper will connect to fetch the data. There should be a windows service exporter listening at that address and port"`
	AllowList   string `default:"" help:"Comma separated list of names of services to be included. By default no service is included"`
	AllowRegex  string `default:"" help:"If set, the Regex specified will be applied to filter in services. es : \"^win\" will include all services starting with \"win\"."`
	DenyList    string `default:"" help:"Comma separated list of names of services to be excluded. This is the last rule applied that take precedence over -allowList and -allowRegex"`
}

const (
	integrationName    = "com.newrelic.winservices"
	integrationVersion = "0.0.2"
)

var (
	args argumentList
)

func main() {
	var metricsByFamily scraper.MetricFamiliesByName

	log.SetupLogging(args.Verbose)

	integrationInstance, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	metricsByFamily, err = scraper.Get(http.DefaultClient, args.ExporterURL)

	validator := nri.NewValidator(args.AllowList, args.DenyList, args.AllowRegex)
	if err := nri.Process(integrationInstance, metricsByFamily, validator); err != nil {
		log.Error(err.Error())
	}

	err = integrationInstance.Publish()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

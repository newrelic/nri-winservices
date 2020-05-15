package main

import (
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/newrelic/nri-winservices/src/nri"
	"github.com/newrelic/nri-winservices/src/scraper"

	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
)

type argumentList struct {
	Verbose     bool   `default:"false" help:"Print more information to logs."`
	Pretty      bool   `default:"false" help:"Print pretty formatted JSON."`
	FakeData    bool   `default:"" help:"The scraper will not connect to a real exporter, but will report fake data"`
	ExporterURL string `default:"http://localhost:9182/metrics" help:"The url to which the scraper will connect to fetch the data. There should be a windows service exporter listening at that address and port"`
	AllowList   string `default:"" help:"Comma separated list of names of services to be included. By default no service is included"`
	AllowRegex  string `default:"" help:"If set, the Regex specified will applied to filter in services. es : \"^win\" will include all services starting with \"win\"."`
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

	if args.FakeData {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "scraper/testdata/actualOutput")
		}))
		defer ts.Close()
		metricsByFamily, err = scraper.Get(http.DefaultClient, ts.URL)
	} else {
		metricsByFamily, err = scraper.Get(http.DefaultClient, args.ExporterURL)
	}
	if err := nri.Process(integrationInstance, metricsByFamily, args.AllowList, args.DenyList, args.AllowRegex); err != nil {
		log.Error(err.Error())

	}
	err = integrationInstance.Publish()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

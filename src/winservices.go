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
	Verbose bool `default:"false" help:"Print more information to logs."`
	Pretty  bool `default:"false" help:"Print pretty formatted JSON."`
}

const (
	integrationName    = "com.newrelic.winservices"
	integrationVersion = "0.0.2"
)

var (
	args argumentList
)

func main() {
	log.SetupLogging(args.Verbose)

	integrationInstance, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "scraper/testdata/actualOutput")
	}))
	defer ts.Close()

	metricsByFamily, err := scraper.Get(http.DefaultClient, ts.URL)

	// metricsByFamily, err := scraper.Get(http.DefaultClient, "http://localhost:9182/metrics")

	if err := nri.Process(integrationInstance, metricsByFamily); err != nil {
		log.Error(err.Error())

	}
	err = integrationInstance.Publish()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

package aws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	docker "github.com/docker/docker/api/types"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"

	"github.com/newrelic/nri-docker/src/raw"
)

const fargateClientTimeout = 30 * time.Second
const fargateTaskStatsCacheKey = "fargate-task-stats"

var fargateHTTPClient = &http.Client{Timeout: fargateClientTimeout}

type timedDockerStats struct {
	docker.Stats
	time time.Time
}

// FargateStats holds a map of Fargate container IDs as key and their Docker metrics
// as values.
type FargateStats map[string]*timedDockerStats

// FargateFetcher fetches metrics from Fargate endpoints in AWS ECS.
type FargateFetcher struct {
	baseURL        *url.URL
	http           *http.Client
	containerStore persist.Storer
	latestFetch    time.Time
}

// NewFargateFetcher creates a new FargateFetcher with the given HTTP client.
func NewFargateFetcher(baseURL *url.URL) (*FargateFetcher, error) {
	containerStore := persist.NewInMemoryStore()

	return &FargateFetcher{
		baseURL:        baseURL,
		http:           fargateHTTPClient,
		containerStore: containerStore,
	}, nil
}

// Fetch fetches raw metrics from a given Fargate container.
func (e *FargateFetcher) Fetch(container docker.ContainerJSON) (raw.Metrics, error) {
	var stats FargateStats
	err := e.fargateStatsFromCacheOrNew(&stats)
	if err != nil {
		return raw.Metrics{}, err
	}
	rawMetrics := fargateRawMetrics(stats)
	return *rawMetrics[container.ID], nil
}

// fargateStatsFromCacheOrNew wraps the access to Fargate task stats with a caching layer.
func (e *FargateFetcher) fargateStatsFromCacheOrNew(response *FargateStats) error {
	defer func() {
		if err := e.containerStore.Save(); err != nil {
			log.Warn("error persisting Fargate task metadata: %s", err)
		}
	}()

	var err error
	_, err = e.containerStore.Get(fargateTaskStatsCacheKey, response)
	if err == persist.ErrNotFound {
		err = e.getFargateContainerMetrics(response)
	}
	if err != nil {
		return fmt.Errorf("cannot fetch task stats response: %s", err)
	}
	e.containerStore.Set(fargateTaskMetadataCacheKey, *response)
	return nil
}

// getFargateContainerMetrics returns Docker metrics from inside a Fargate container.
// It captures the ECS container metadata endpoint from the environment variable defined by
// `containerMetadataEnvVar`.
// Note that the endpoint doesn't follow strictly the same schema as Docker's: it returns a list of containers,
// instead of only one. They are not compatible in terms of the requests that they accept, but they share
// part of the response's schema.
func (e *FargateFetcher) getFargateContainerMetrics(stats *FargateStats) error {
	endpoint := TaskStatsEndpoint(e.baseURL.String())

	response, err := metadataResponse(e.http, endpoint)
	if err != nil {
		return fmt.Errorf(
			"error when sending request to ECS container metadata endpoint (%s): %v",
			endpoint,
			err,
		)
	}
	log.Debug("fargate task stats response from endpoint %s: %s", endpoint, string(response))
	e.latestFetch = time.Now()

	err = json.Unmarshal(response, &stats)
	if err != nil {
		return fmt.Errorf("error unmarshalling ECS container: %v", err)
	}

	now := time.Now()
	for id := range *stats {
		(*stats)[id].time = now
	}

	return nil
}

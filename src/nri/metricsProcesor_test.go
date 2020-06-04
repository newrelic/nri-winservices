package nri

import (
	"errors"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/nri-winservices/src/matcher"
	"github.com/newrelic/nri-winservices/src/scraper"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	serviceName        = "rpcss"
	serviceStartMode   = "auto"
	serviceDisplayName = "Remote Procedure Call (RPC)"
	servicePid         = "668"
	hostname           = "test-hostname"
)

var gauge = dto.MetricType_GAUGE

var metricFamlilyServiceInfo = dto.MetricFamily{
	Name: strPtr("windows_service_info"),
	Type: &gauge,
	Metric: []*dto.Metric{
		{
			Label: []*dto.LabelPair{
				{
					Name:  strPtr("name"),
					Value: strPtr(serviceName),
				},
				{
					Name:  strPtr("display_name"),
					Value: strPtr(serviceDisplayName),
				},
				{
					Name:  strPtr("process_id"),
					Value: strPtr(servicePid),
				},
			},
			Gauge: &dto.Gauge{
				Value: float64Ptr(1),
			},
		},
	},
}
var metricFamlilyService = dto.MetricFamily{
	Name: strPtr("windows_service_start_mode"),
	Type: &gauge,
	Metric: []*dto.Metric{
		{
			Label: []*dto.LabelPair{
				{
					Name:  strPtr("name"),
					Value: strPtr(serviceName),
				},
				{
					Name:  strPtr("start_mode"),
					Value: strPtr(serviceStartMode),
				},
			},
			Gauge: &dto.Gauge{
				Value: float64Ptr(1),
			},
		},
		{
			Label: []*dto.LabelPair{
				{
					Name:  strPtr("name"),
					Value: strPtr("rpcss"),
				},
				{
					Name:  strPtr("start_mode"),
					Value: strPtr("boot"),
				},
			},
			Gauge: &dto.Gauge{
				Value: float64Ptr(0),
			},
		},
	},
}
var metricFamlilyServiceHostname = dto.MetricFamily{
	Name: strPtr("windows_cs_hostname"),
	Type: &gauge,
	Metric: []*dto.Metric{
		{
			Label: []*dto.LabelPair{
				{
					Name:  strPtr("hostname"),
					Value: strPtr(hostname),
				},
			},
			Gauge: &dto.Gauge{
				Value: float64Ptr(1),
			},
		},
	},
}

func TestCreateEntities(t *testing.T) {
	i, _ := integration.New("integrationName", "integrationVersion")
	rules := loadRules()
	mfbn := scraper.MetricFamiliesByName{
		"windows_service_info":       metricFamlilyServiceInfo,
		"windows_service_start_mode": metricFamlilyService,
		"windows_cs_hostname":        metricFamlilyServiceHostname,
	}

	h, err := getHostname(mfbn, rules)
	matcher := matcher.New(`"` + serviceName + `"`)
	entityMap, err := createEntities(i, mfbn, rules, matcher, h)
	require.NoError(t, err)
	_, ok := entityMap[serviceName]
	require.True(t, ok)
	require.Len(t, i.Entities, 1)
	require.Equal(t, i.Entities[0].Name(), hostname+":"+serviceName)
}

func TestGetHostnameFail(t *testing.T) {
	rules := loadRules()
	mfbn := scraper.MetricFamiliesByName{
		"windows_service_info":       metricFamlilyServiceInfo,
		"windows_service_start_mode": metricFamlilyService,
		// exclude host name metrics from family
		// "windows_cs_hostname":        metricFamlilyServiceHostname,
	}
	_, err := getHostname(mfbn, rules)
	require.Equal(t, err, errors.New("hostname Metric not found"))
}

func TestNoServiceNameAllowed(t *testing.T) {
	i, _ := integration.New("integrationName", "integrationVersion")
	rules := loadRules()
	mfbn := scraper.MetricFamiliesByName{
		"windows_service_info":       metricFamlilyServiceInfo,
		"windows_service_start_mode": metricFamlilyService,
		"windows_cs_hostname":        metricFamlilyServiceHostname,
	}

	matcher := matcher.New("")
	h, err := getHostname(mfbn, rules)
	entityMap, err := createEntities(i, mfbn, rules, matcher, h)
	require.NoError(t, err, "No error is expected even if no service is allowed")
	require.Len(t, entityMap, 0, "No entity is expected since no service is allowed")
	err = processMetricGauge(metricFamlilyService, rules, entityMap, h)
	err = processMetricGauge(metricFamlilyService, rules, entityMap, h)
	require.NoError(t, err)
	require.NoError(t, err, "No error is expected even if entityMap is empty")
}

func TestProccessMetricGauge(t *testing.T) {
	i, _ := integration.New("integrationName", "integrationVersion")
	rules := loadRules()
	mfbn := scraper.MetricFamiliesByName{
		"windows_service_info":       metricFamlilyServiceInfo,
		"windows_service_start_mode": metricFamlilyService,
		"windows_cs_hostname":        metricFamlilyServiceHostname,
	}

	h, err := getHostname(mfbn, rules)
	matcher := matcher.New(`"` + serviceName + `"`)
	entityMap, err := createEntities(i, mfbn, rules, matcher, h)
	require.NoError(t, err)
	// process info metrics
	err = processMetricGauge(metricFamlilyServiceInfo, rules, entityMap, h)
	require.NoError(t, err)
	metadata := entityMap[serviceName].GetMetadata()
	assert.Equal(t, serviceDisplayName, metadata["windowsService.displayName"])
	assert.Equal(t, servicePid, metadata["windowsService.processId"])
	// process startmode metrics
	err = processMetricGauge(metricFamlilyService, rules, entityMap, h)
	assert.NoError(t, err)
	assert.Equal(t, serviceStartMode, metadata["windowsService.startMode"])
	assert.Equal(t, hostname, metadata[rules.EntityName.HostnameNrdbLabelName])
	assert.Equal(t, hostname+":"+serviceName, metadata[rules.EntityName.EntityNrdbLabelName])

}

func strPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

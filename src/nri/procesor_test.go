package nri

import (
	"errors"
	"github.com/newrelic/nri-winservices/src/scraper"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newrelic/infra-integrations-sdk/integration"
	dto "github.com/prometheus/client_model/go"
)

var gauge = dto.MetricType_GAUGE
var metricFamlilyService = dto.MetricFamily{
	Name: strPtr("wmi_service_start_mode"),
	Type: &gauge,
	Metric: []*dto.Metric{
		{
			Label: []*dto.LabelPair{
				{
					Name:  strPtr("name"),
					Value: strPtr("rpcss"),
				},
				{
					Name:  strPtr("start_mode"),
					Value: strPtr("auto"),
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
	Name: strPtr("wmi_cs_hostname"),
	Type: &gauge,
	Metric: []*dto.Metric{
		{
			Label: []*dto.LabelPair{
				{
					Name:  strPtr("hostname"),
					Value: strPtr("test-hostname"),
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
	mfbn := scraper.MetricFamiliesByName{"wmi_service_start_mode": metricFamlilyService, "wmi_cs_hostname": metricFamlilyServiceHostname}

	entityMap, err := createEntities(i, mfbn, rules)

	assert.Nil(t, err)
	_, ok := entityMap["rpcss"]
	assert.True(t, ok)

	assert.Equal(t, i.Entities[0].Name(), "test-hostname:rpcss")

}

func TestCreateEntitiesFail(t *testing.T) {
	i, _ := integration.New("integrationName", "integrationVersion")
	rules := loadRules()
	mfbn := scraper.MetricFamiliesByName{"wmi_service_start_mode": metricFamlilyService}
	_, err := createEntities(i, mfbn, rules)

	assert.Equal(t, err, errors.New("hostName Metric not found"))
}

func TestProccessMetricGauge(t *testing.T) {
	i, _ := integration.New("integrationName", "integrationVersion")
	rules := loadRules()
	mfbn := scraper.MetricFamiliesByName{"wmi_service_start_mode": metricFamlilyService, "wmi_cs_hostname": metricFamlilyServiceHostname}

	entityMap, err := createEntities(i, mfbn, rules)
	assert.Nil(t, err)
	err = processMetricGauge(metricFamlilyService, rules, entityMap)
	assert.NoError(t, err)

}

func strPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

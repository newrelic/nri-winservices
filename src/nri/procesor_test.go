package nri

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newrelic/infra-integrations-sdk/integration"
	dto "github.com/prometheus/client_model/go"
)

var metricFamlily = dto.MetricFamily{
	Name: &(&struct{ x string }{"wmi_service_start_mode"}).x,
	Type: &(&struct{ x dto.MetricType }{dto.MetricType_GAUGE}).x,
	Metric: []*dto.Metric{
		{
			Label: []*dto.LabelPair{
				{
					Name:  &(&struct{ x string }{"name"}).x,
					Value: &(&struct{ x string }{"rpcss"}).x,
				},
				{
					Name:  &(&struct{ x string }{"start_mode"}).x,
					Value: &(&struct{ x string }{"auto"}).x,
				},
			},
			Gauge: &dto.Gauge{
				Value: &(&struct{ x float64 }{1}).x,
			},
		},
		{
			Label: []*dto.LabelPair{
				{
					Name:  &(&struct{ x string }{"name"}).x,
					Value: &(&struct{ x string }{"rpcss"}).x,
				},
				{
					Name:  &(&struct{ x string }{"start_mode"}).x,
					Value: &(&struct{ x string }{"boot"}).x,
				},
			},
			Gauge: &dto.Gauge{
				Value: &(&struct{ x float64 }{0}).x,
			},
		},
	},
}

func TestCreateEntities(t *testing.T) {
	i, _ := integration.New("integrationName", "integrationVersion")
	rules := &loadRules("testdata/rulesTest.yml").EntityRules[0]
	ebn := createEntities(i, metricFamlily, rules)
	_, ok := ebn["rpcss"]
	assert.True(t, ok) //todo improve the assert
}
func TestProccessMetricGauge(t *testing.T) {
	i, _ := integration.New("integrationName", "integrationVersion")

	rules := &loadRules("testdata/rulesTest.yml").EntityRules[0]
	ebn := createEntities(i, metricFamlily, rules)

	err := processMetricGauge(metricFamlily, rules, ebn)
	assert.NoError(t, err)
	// i.Publish()

}

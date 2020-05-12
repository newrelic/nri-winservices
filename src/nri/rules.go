package nri

import (
	"fmt"
)

type EntityRules struct {
	EntityType string        `yaml:"type"`
	EntityName EntityName    `yaml:"name"`
	Metrics    []MetricRules `yaml:"metrics"`
}
type EntityName struct {
	Metric              string `yaml:"from_metric"`
	MetricLabel         string `yaml:"use_label"`
	HostNameMetric      string
	HostNameMetricLabel string
}
type MetricRules struct {
	ProviderName string   `yaml:"provider_name"`
	MetricType   string   `yaml:"type"`
	NrdbName     string   `yaml:"nrdb_name"`
	SkipValue    float64  `yaml:"skip_value"`
	Attributes   []string `yaml:"attributes"`
}

func loadRules() EntityRules {

	rules := EntityRules{

		EntityType: "WindowsService",
		EntityName: EntityName{
			Metric:              "wmi_service_start_mode",
			MetricLabel:         "name",
			HostNameMetric:      "wmi_cs_hostname",
			HostNameMetricLabel: "hostname",
		},
		Metrics: []MetricRules{
			{
				ProviderName: "wmi_service_start_mode",
				MetricType:   "gauge",
				NrdbName:     "startMode",
				SkipValue:    0,
				Attributes:   []string{"start_mode"},
			},
			{
				ProviderName: "wmi_service_state",
				MetricType:   "gauge",
				NrdbName:     "state",
				SkipValue:    0,
				Attributes:   []string{"state"},
			},
			{
				ProviderName: "wmi_service_status",
				MetricType:   "gauge",
				NrdbName:     "status",
				SkipValue:    0,
				Attributes:   []string{"status"},
			},
		},
	}

	return rules
}

func (r *EntityRules) getMetricRules(providerName string) (*MetricRules, error) {
	for _, m := range r.Metrics {
		if m.ProviderName == providerName {
			return &m, nil //todo check if copy
		}
	}
	return nil, fmt.Errorf("No rules find for providerName: %s", providerName)
}

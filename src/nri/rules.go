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
	HostNameMetric      string `yaml:"hostname_metric"`
	HostNameMetricLabel string `yaml:"hostname_label"`
}
type MetricRules struct {
	ProviderName string      `yaml:"provider_name"`
	MetricType   string      `yaml:"type"`
	NrdbName     string      `yaml:"nrdb_name"`
	SkipValue    float64     `yaml:"skip_value"`
	Attributes   []Attribute `yaml:"attributes"`
}
type Attribute struct {
	Label            string `yaml:"provider_name"`
	NrdbLabelName    string `yaml:"nrdb_name"`
	IsEntityMetadata bool   `yaml:"entity_attribute"`
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
				NrdbName:     "windowsService.startMode",
				SkipValue:    0,
				Attributes: []Attribute{
					{
						Label:            "start_mode",
						NrdbLabelName:    "windowsService.startModeLabel",
						IsEntityMetadata: true,
					},
					{
						Label:            "name",
						NrdbLabelName:    "windowsService.name",
						IsEntityMetadata: true,
					},
				},
			},
			{
				ProviderName: "wmi_service_state",
				MetricType:   "gauge",
				NrdbName:     "windowsService.state",
				SkipValue:    0,
				Attributes: []Attribute{
					{
						Label:         "state",
						NrdbLabelName: "windowsService.stateLabel",
					},
					{
						Label:            "name",
						NrdbLabelName:    "windowsService.name",
						IsEntityMetadata: true,
					},
				},
			},
			{
				ProviderName: "wmi_service_status",
				MetricType:   "gauge",
				NrdbName:     "windowsService.status",
				SkipValue:    0,
				Attributes: []Attribute{
					{
						Label:         "status",
						NrdbLabelName: "windowsService.statusLabel",
					},
					{
						Label:            "name",
						NrdbLabelName:    "windowsService.name",
						IsEntityMetadata: true,
					},
				},
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
	return nil, fmt.Errorf("no rules find for providerName: %s", providerName)
}

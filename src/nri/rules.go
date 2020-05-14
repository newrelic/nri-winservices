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
	Metric                string `yaml:"from_metric"`
	MetricLabel           string `yaml:"use_label"`
	HostnameMetric        string `yaml:"hostname_metric"`
	HostnameMetricLabel   string `yaml:"hostname_label"`
	HostnameNrdbLabelName string `yaml:"hostname_nrdb_name"`
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
			Metric:                "wmi_service_start_mode",
			MetricLabel:           "name",
			HostnameMetric:        "wmi_cs_hostname",
			HostnameMetricLabel:   "hostname",
			HostnameNrdbLabelName: "windowsService.hostname",
		},
		Metrics: []MetricRules{
			{
				ProviderName: "wmi_service_start_mode",
				MetricType:   "gauge",
				NrdbName:     "windowsService.service.startMode",
				SkipValue:    0,
				Attributes: []Attribute{
					{
						Label:            "start_mode",
						NrdbLabelName:    "windowsService.startMode",
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
				NrdbName:     "windowsService.service.state",
				SkipValue:    0,
				Attributes: []Attribute{
					{
						Label:         "state",
						NrdbLabelName: "windowsService.state",
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
				NrdbName:     "windowsService.service.status",
				SkipValue:    0,
				Attributes: []Attribute{
					{
						Label:         "status",
						NrdbLabelName: "windowsService.status",
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

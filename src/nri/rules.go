package nri

import (
	"fmt"
)

// EntityRules represents rules to convert prometheus metrics into NewRelic entities.
type EntityRules struct {
	EntityType string        `yaml:"type"`
	EntityName EntityName    `yaml:"name"`
	Metrics    []MetricRules `yaml:"metrics"`
}

// EntityName indicates which metrics labels use to form the unique entity name and displayName
// for windows services the entity name is hostname:serviceName.
type EntityName struct {
	Metric                string `yaml:"from_metric"`
	Label                 string `yaml:"name_label"`
	DisplayNameLabel      string `yaml:"display_name_label"`
	EntityNrdbLabelName   string `yaml:"entity_nrdb_name"`
	HostnameMetric        string `yaml:"hostname_metric"`
	HostnameLabel         string `yaml:"hostname_label"`
	HostnameNrdbLabelName string `yaml:"hostname_nrdb_name"`
}

// MetricRules describe the metrics that compose the entity.
//
// prometheus enums metrics are generally send using following style:
// 	wmi_service_start_mode{name="wersvc",start_mode="auto"} 0
// 	wmi_service_start_mode{name="wersvc",start_mode="disabled"} 0
// 	wmi_service_start_mode{name="wersvc",start_mode="manual"} 1
// using EnumMetric=true will only send the metric with value 1 with the corresponding attribute
//
// for promethus *_info metrics no metric will be send, just metadata.
type MetricRules struct {
	ProviderName string      `yaml:"provider_name"`
	MetricType   string      `yaml:"type"`
	NrdbName     string      `yaml:"nrdb_name"`
	EnumMetric   bool        `yaml:"enum_metric"`
	InfoMetric   bool        `yaml:"info_metric"`
	Attributes   []Attribute `yaml:"attributes"`
}

// Attribute describe metrics attributes to be add.
type Attribute struct {
	Label            string `yaml:"provider_name"`
	NrdbLabelName    string `yaml:"nrdb_name"`
	IsEntityMetadata bool   `yaml:"entity_metadata"` // when true this attribute will be use as metadata.
}

func loadRules() EntityRules {

	rules := EntityRules{

		EntityType: "WindowsService",
		EntityName: EntityName{
			Metric:                "wmi_service_info",
			Label:                 "name",
			DisplayNameLabel:      "display_name",
			EntityNrdbLabelName:   "windowsService.entityName",
			HostnameMetric:        "wmi_cs_hostname",
			HostnameLabel:         "hostname",
			HostnameNrdbLabelName: "windowsService.hostname",
		},
		Metrics: []MetricRules{
			{
				ProviderName: "wmi_service_info",
				InfoMetric:   true,
				Attributes: []Attribute{
					{
						Label:            "name",
						NrdbLabelName:    "windowsService.name",
						IsEntityMetadata: true,
					},
					{
						Label:            "display_name",
						NrdbLabelName:    "windowsService.displayName",
						IsEntityMetadata: true,
					},
					{
						Label:            "process_id",
						NrdbLabelName:    "windowsService.processId",
						IsEntityMetadata: true,
					},
				},
			},
			{
				ProviderName: "wmi_service_start_mode",
				MetricType:   "gauge",
				NrdbName:     "windowsService.service.startMode",
				EnumMetric:   true,
				Attributes: []Attribute{
					{
						Label:            "start_mode",
						NrdbLabelName:    "windowsService.startMode",
						IsEntityMetadata: true,
					},
				},
			},
			{
				ProviderName: "wmi_service_state",
				MetricType:   "gauge",
				NrdbName:     "windowsService.service.state",
				EnumMetric:   true,
				Attributes: []Attribute{
					{
						Label:         "state",
						NrdbLabelName: "windowsService.state",
					},
				},
			},
			{
				ProviderName: "wmi_service_status",
				MetricType:   "gauge",
				NrdbName:     "windowsService.service.status",
				EnumMetric:   true,
				Attributes: []Attribute{
					{
						Label:         "status",
						NrdbLabelName: "windowsService.status",
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

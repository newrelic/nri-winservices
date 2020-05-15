package nri

import (
	"fmt"
	"time"

	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-winservices/src/scraper"
	dto "github.com/prometheus/client_model/go"
)

//This constant is needed only till the workaround to register entity is in place
const entityTypeInventory = "windowsService"

type entitiesByName map[string]*integration.Entity
type metadataMap map[string]string
type attributesMap map[string]string

// Process creates entities and add metrics from the MetricFamiliesByName according to rules
func Process(i *integration.Integration, metricFamilyMap scraper.MetricFamiliesByName, validator Validator) error {
	entityRules := loadRules()

	entityMap, err := createEntities(i, metricFamilyMap, entityRules, validator)
	if err != nil {
		return err
	}

	for _, metricsRules := range entityRules.Metrics {
		if metricFamily, ok := metricFamilyMap[metricsRules.ProviderName]; ok {
			if err := processMetricGauge(metricFamily, entityRules, entityMap); err != nil {
				log.Warn("error processing metric: %v", err.Error())
			}
		}
	}
	return nil
}

func createEntities(integrationInstance *integration.Integration, metricFamilyMap scraper.MetricFamiliesByName, entityRules EntityRules, validator Validator) (entitiesByName, error) {
	entityMap := make(map[string]*integration.Entity)

	mfHostname, ok := metricFamilyMap[entityRules.EntityName.HostnameMetric]
	if !ok {
		return nil, fmt.Errorf("hostname Metric not found")
	}

	hostname, err := getHostname(mfHostname, entityRules)
	if err != nil {
		return nil, err
	}

	mf, ok := metricFamilyMap[entityRules.EntityName.Metric]
	if !ok {
		return nil, fmt.Errorf("entityName Metric not found")
	}
	for _, metric := range mf.GetMetric() {
		serviceName, err := getLabelValue(metric.GetLabel(), entityRules.EntityName.Label)
		if err != nil {
			warnOnErr(err)
			continue
		}

		shouldBeIncluded := validator.ValidateServiceName(serviceName)

		if !shouldBeIncluded {
			continue
		}

		if _, ok := entityMap[serviceName]; ok {
			continue
		}
		serviceDisplayName, err := getLabelValue(metric.GetLabel(), entityRules.EntityName.DisplayNameLabel)
		if err != nil {
			warnOnErr(err)
			continue
		}
		entityName := hostname + ":" + serviceName

		entity, err := integrationInstance.NewEntity(entityName, entityRules.EntityType, serviceDisplayName)
		if err != nil {
			warnOnErr(err)
			continue
		}
		integrationInstance.AddEntity(entity)
		err = entity.AddInventoryItem(entityTypeInventory, "name", entityName)
		warnOnErr(err)
		err = entity.AddInventoryItem(entityTypeInventory, entityRules.EntityName.HostnameNrdbLabelName, hostname)
		warnOnErr(err)

		entityMap[serviceName] = entity
	}
	return entityMap, nil
}

func processMetricGauge(metricFamily dto.MetricFamily, entityRules EntityRules, ebn entitiesByName) error {
	metricRules, err := entityRules.getMetricRules(metricFamily.GetName())
	if err != nil {
		return fmt.Errorf("metric rule not found")
	}
	if metricFamily.GetType() != dto.MetricType_GAUGE {
		return fmt.Errorf("metric type not Gauge")
	}
	for _, metric := range metricFamily.GetMetric() {
		metricValue := metric.GetGauge().GetValue()
		// skip enum metrics without value
		if metricRules.EnumMetric && metricValue != 1 {
			continue
		}

		serviceName, err := getLabelValue(metric.GetLabel(), entityRules.EntityName.Label)
		if err != nil {
			return err
		}
		e, ok := ebn[serviceName]
		if !ok {
			continue
		}

		attributes, metadata := getAttributesAndMetadata(metricRules.Attributes, metric)
		addMetadata(metadata, e)
		// _info metrics only contains metadata
		if metricRules.InfoMetric {
			continue
		}

		metricName := metricRules.NrdbName
		gauge, err := integration.Gauge(time.Now(), metricName, metricValue)
		warnOnErr(err)
		for k, v := range attributes {
			err = gauge.AddDimension(k, v)
			warnOnErr(err)
		}
		e.AddMetric(gauge)
	}
	return nil
}

func addMetadata(metadata metadataMap, e *integration.Entity) {
	var err error
	for k, v := range metadata {
		err = e.AddMetadata(k, v)
		warnOnErr(err)
		err = e.AddInventoryItem(entityTypeInventory, k, v)
		warnOnErr(err)
	}
}

func getAttributesAndMetadata(attributesRules []Attribute, metric *dto.Metric) (attributesMap, metadataMap) {
	var metadata = make(map[string]string)
	var attributes = make(map[string]string)
	for _, attribute := range attributesRules {
		value, err := getLabelValue(metric.GetLabel(), attribute.Label)
		if err != nil {
			log.Warn(err.Error())
			continue
		}
		nrdbLabelName := attribute.NrdbLabelName

		if attribute.IsEntityMetadata {
			metadata[nrdbLabelName] = value
			continue
		}
		attributes[nrdbLabelName] = value
	}
	return attributes, metadata
}

func getLabelValue(label []*dto.LabelPair, key string) (string, error) {
	for _, l := range label {
		if l.GetName() == key {
			return l.GetValue(), nil
		}
	}
	return "", fmt.Errorf("label %v not found", key)
}

func getHostname(mf dto.MetricFamily, entityRules EntityRules) (string, error) {
	var hostname string
	var err error
	for _, m := range mf.GetMetric() {
		hostname, err = getLabelValue(m.GetLabel(), entityRules.EntityName.HostnameLabel)
		if err != nil {
			return "", err
		}
	}
	return hostname, nil
}

func warnOnErr(err error) {
	if err != nil {
		log.Warn(err.Error())
	}
}

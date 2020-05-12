package nri

import (
	"fmt"
	"time"

	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-winservices/src/scraper"
	dto "github.com/prometheus/client_model/go"
)

type entitiesByName map[string]*integration.Entity

// Process creates entities and add metrics from the MetricFamiliesByName according to rules
func Process(i *integration.Integration, metricFamilyMap scraper.MetricFamiliesByName) error {
	entityRules := loadRules()

	entityMap, err := createEntities(i, metricFamilyMap, entityRules)
	if err != nil {
		return err
	}

	for _, metricsRules := range entityRules.Metrics {
		if metricFamily, ok := metricFamilyMap[metricsRules.ProviderName]; ok {
			if err := processMetricGauge(metricFamily, entityRules, entityMap); err != nil {
				return err
			}
		}
	}
	return nil
}

func createEntities(integrationInstance *integration.Integration, metricFamilyMap scraper.MetricFamiliesByName, entityRules EntityRules) (entitiesByName, error) {
	entityMap := make(map[string]*integration.Entity)

	mf, ok := metricFamilyMap[entityRules.EntityName.HostNameMetric]
	if !ok {
		return nil, fmt.Errorf("HostName Metric not found")
	}

	var hostName string
	for _, m := range mf.GetMetric() {
		hostName = getLabelValue(m.GetLabel(), entityRules.EntityName.HostNameMetricLabel)
	}

	mf, ok = metricFamilyMap[entityRules.EntityName.Metric]
	if !ok {
		return nil, fmt.Errorf("EntityName Metric not found")
	}
	for _, metric := range mf.GetMetric() {

		serviceName := getLabelValue(metric.GetLabel(), entityRules.EntityName.MetricLabel)
		if _, ok := entityMap[serviceName]; ok {
			continue
		}
		entityName := hostName + ":" + serviceName

		entity, err := integrationInstance.NewEntity(entityName, entityRules.EntityType, serviceName)
		fatalOnErr(err)
		integrationInstance.AddEntity(entity)
		//todo add metadata
		entityMap[serviceName] = entity
	}
	return entityMap, nil
}

func processMetricGauge(metricFamily dto.MetricFamily, entityRules EntityRules, ebn entitiesByName) error {
	metricRules, err := entityRules.getMetricRules(metricFamily.GetName())
	if err != nil {
		return fmt.Errorf("Metric rule not found")
	}
	if metricFamily.GetType() != dto.MetricType_GAUGE {
		return fmt.Errorf("Metric type not Gauge")
	}
	for _, metric := range metricFamily.GetMetric() {
		metricValue := metric.GetGauge().GetValue()
		if metricValue == metricRules.SkipValue {
			continue
		}

		serviceName := getLabelValue(metric.GetLabel(), entityRules.EntityName.MetricLabel)
		e, ok := ebn[serviceName]
		if !ok {
			log.Error("Entity not found for service: %v", serviceName)
			continue
		}

		metricName := metricRules.NrdbName
		gauge, err := integration.Gauge(time.Now(), metricName, metricValue)
		if err != nil {
			return err
		}

		metricDimension := metricRules.Attributes[0] //todo 1 attribute
		metricDimensionValue := getLabelValue(metric.GetLabel(), metricDimension)
		gauge.AddDimension(metricDimension, metricDimensionValue)
		e.AddMetric(gauge)
	}
	return nil
}

// todo return err and check nil
func getLabelValue(label []*dto.LabelPair, key string) string {
	for _, l := range label {
		if l.GetName() == key {
			return l.GetValue()
		}
	}
	return ""
}

func fatalOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

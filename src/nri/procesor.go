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
	rules := loadRules()
	//todo believe the rule was no needed as a pointer, but maybe you used it for some reason
	entityRules := rules.EntityRules[0] //TODO we should check the length og the arrray

	var entityMap entitiesByName
	//TODO we should check the length og the arrray
	metricEntityBasedName := entityRules.Metrics[0].ProviderName //todo take this from config

	if metricFamily, ok := metricFamilyMap[metricEntityBasedName]; ok {
		entityMap = createEntities(i, metricFamily, entityRules)
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

func createEntities(integrationInstance *integration.Integration, metricFamily dto.MetricFamily, entityRules EntityRules) entitiesByName {
	entityMap := make(map[string]*integration.Entity)
	for _, metric := range metricFamily.GetMetric() {
		// todo attach host entity to name to uniqueness
		serviceName := getLabelValue(metric.GetLabel(), entityRules.Name)

		if _, ok := entityMap[serviceName]; ok {
			continue
		}

		entity, err := integrationInstance.NewEntity(serviceName, entityRules.EntityType, serviceName)
		fatalOnErr(err)
		integrationInstance.AddEntity(entity)
		//todo add metadata
		entityMap[serviceName] = entity
	}
	return entityMap
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

		serviceName := getLabelValue(metric.GetLabel(), entityRules.Name)
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

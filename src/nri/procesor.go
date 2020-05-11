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

// Process creates entities and add metrics from the MetricFamiliesByName acording to rules
func Process(i *integration.Integration, mfn scraper.MetricFamiliesByName) error {
	rules := &loadRules("rules.yml").EntityRules[0]

	var ebn entitiesByName
	metricEntityBasedName := rules.Metrics[0].ProviderName //todo take this from config

	if mf, ok := mfn[metricEntityBasedName]; ok {
		ebn = createEntities(i, mf, rules)
	}

	for _, metricsRules := range rules.Metrics {
		if mf, ok := mfn[metricsRules.ProviderName]; ok {
			if err := processMetricGauge(mf, rules, ebn); err != nil {
				return err
			}
		}
	}
	return nil
}

func createEntities(i *integration.Integration, mf dto.MetricFamily, rules *EntityRules) entitiesByName {
	ebn := make(map[string]*integration.Entity)
	for _, m := range mf.GetMetric() {
		// todo attach hostentity to name to uniqueness
		serviceName := getLabelValue(m.GetLabel(), rules.Name)

		if _, ok := ebn[serviceName]; ok {
			continue
		}

		e, err := i.NewEntity(serviceName, rules.EntityType, serviceName)
		fatalOnErr(err)
		i.AddEntity(e)
		//todo add metadata
		ebn[serviceName] = e
	}
	return ebn
}

func processMetricGauge(mf dto.MetricFamily, rules *EntityRules, ebn entitiesByName) error {
	metricRules, err := rules.getMetricRules(mf.GetName())
	if err != nil {
		return fmt.Errorf("Metric rule not found")
	}
	if mf.GetType() != dto.MetricType_GAUGE {
		return fmt.Errorf("Metric type not Gauge")
	}
	for _, metric := range mf.GetMetric() {
		metricValue := metric.GetGauge().GetValue()
		if metricValue == metricRules.SkipValue {
			continue
		}

		serviceName := getLabelValue(metric.GetLabel(), rules.Name)
		e, ok := ebn[serviceName]
		if !ok {
			log.Error("Entity not found for service: %v", serviceName)
			continue
		}

		metricName := metricRules.NrdbName
		g, err := integration.Gauge(time.Now(), metricName, metricValue)
		if err != nil {
			return err
		}

		metricDimension := metricRules.Attributes[0] //todo 1 attribute
		metricDimensionValue := getLabelValue(metric.GetLabel(), metricDimension)
		g.AddDimension(metricDimension, metricDimensionValue)
		e.AddMetric(g)
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

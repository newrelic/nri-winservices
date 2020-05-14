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

		serviceName, err := getLabelValue(metric.GetLabel(), entityRules.EntityName.MetricLabel)
		if err != nil {
			return nil, err
		}
		if _, ok := entityMap[serviceName]; ok {
			continue
		}
		entityName := hostname + ":" + serviceName

		entity, err := integrationInstance.NewEntity(entityName, entityRules.EntityType, serviceName)
		fatalOnErr(err)
		integrationInstance.AddEntity(entity)
		err = entity.AddInventoryItem(entityTypeInventory, "name", entityName)
		if err != nil {
			log.Warn(err.Error())
		}
		err = entity.AddInventoryItem(entityTypeInventory, entityRules.EntityName.HostnameNrdbLabelName, hostname)
		if err != nil {
			log.Warn(err.Error())
		}

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
		if metricValue == metricRules.SkipValue {
			continue
		}

		serviceName, err := getLabelValue(metric.GetLabel(), entityRules.EntityName.MetricLabel)
		if err != nil {
			return err
		}
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
		// Add Metrics attributes
		for _, attribute := range metricRules.Attributes {
			value, err := getLabelValue(metric.GetLabel(), attribute.Label)
			if err != nil {
				return err
			}
			nrdbLabelName := attribute.NrdbLabelName
			err = gauge.AddDimension(nrdbLabelName, value)
			if err != nil {
				log.Warn(err.Error())
			}
			// Add entity metadata for attributes
			if attribute.IsEntityMetadata {
				err = e.AddMetadata(nrdbLabelName, value)
				if err != nil {
					log.Warn(err.Error())
				}
				err = e.AddInventoryItem(entityTypeInventory, nrdbLabelName, value)
				if err != nil {
					log.Warn(err.Error())
				}
			}
		}
		// TODO Remove this when metadata decoration is available for DM.
		for k, v := range e.GetMetadata() {
			_ = gauge.AddDimension(k, v.(string))
		}
		e.AddMetric(gauge)
	}
	return nil
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
		hostname, err = getLabelValue(m.GetLabel(), entityRules.EntityName.HostnameMetricLabel)
		if err != nil {
			return "", err
		}
	}
	return hostname, nil
}

func fatalOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

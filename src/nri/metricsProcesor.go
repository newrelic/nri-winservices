/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package nri

import (
	"fmt"
	"time"

	"github.com/newrelic/nri-winservices/src/matcher"

	"github.com/newrelic/infra-integrations-sdk/v4/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v4/integration"
	"github.com/newrelic/infra-integrations-sdk/v4/log"
	"github.com/newrelic/nri-winservices/src/scraper"
	dto "github.com/prometheus/client_model/go"
)

const (
	entityNamePrefix = "WIN_SERVICE"
	// localhost will be automatically replaced by the host_name or display_name by the agent when is found inside EntityName
	hostName = "localhost"
)

type entitiesByName map[string]*integration.Entity
type metadataMap map[string]string
type attributesMap map[string]string

// ProcessMetrics creates entities and add metrics from the MetricFamiliesByName according to rules
func ProcessMetrics(i *integration.Integration, metricFamilyMap scraper.MetricFamiliesByName, matcher matcher.Matcher, hostname string) error {
	entityRules := loadRules()

	if hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}

	entityMap, err := createEntities(i, metricFamilyMap, entityRules, matcher)
	if err != nil {
		return err
	}

	for _, metricsRules := range entityRules.Metrics {
		if metricFamily, ok := metricFamilyMap[metricsRules.ProviderName]; ok {
			if err := processMetricGauge(metricFamily, entityRules, entityMap, hostname); err != nil {
				log.Warn("error processing metric:%v", err.Error())
			}
		}
	}
	return nil
}

func createEntities(integrationInstance *integration.Integration, metricFamilyMap scraper.MetricFamiliesByName, entityRules EntityRules, matcher matcher.Matcher) (entitiesByName, error) {
	entityMap := make(map[string]*integration.Entity)

	mf, ok := metricFamilyMap[entityRules.EntityName.Metric]
	if !ok {
		return nil, fmt.Errorf("entityName Metric not found")
	}
	for _, m := range mf.GetMetric() {
		serviceName, err := getLabelValue(m.GetLabel(), entityRules.EntityName.Label)
		if err != nil {
			warnOnErr(err)
			continue
		}

		shouldBeIncluded := matcher.Match(serviceName)

		if !shouldBeIncluded {
			continue
		}

		if _, ok := entityMap[serviceName]; ok {
			continue
		}
		serviceDisplayName, err := getLabelValue(m.GetLabel(), entityRules.EntityName.DisplayNameLabel)
		if err != nil {
			warnOnErr(err)
			continue
		}

		entityName := fmt.Sprintf("%s:%s:%s", entityNamePrefix, hostName, serviceName)

		entity, err := integrationInstance.NewEntity(entityName, entityRules.EntityType, serviceDisplayName)
		if err != nil {
			warnOnErr(err)
			continue
		}
		integrationInstance.AddEntity(entity)

		entityMap[serviceName] = entity
	}
	return entityMap, nil
}

func processMetricGauge(metricFamily dto.MetricFamily, entityRules EntityRules, ebn entitiesByName, hostname string) error {
	metricRules, err := entityRules.getMetricRules(metricFamily.GetName())
	if err != nil {
		return fmt.Errorf("metric rule not found")
	}
	if metricFamily.GetType() != dto.MetricType_GAUGE {
		return fmt.Errorf("metric type not Gauge")
	}
	noMetricAdded := true
	for _, m := range metricFamily.GetMetric() {
		metricValue := m.GetGauge().GetValue()
		// skip enum metrics without value
		if metricRules.EnumMetric && metricValue != 1 {
			continue
		}

		serviceName, err := getLabelValue(m.GetLabel(), entityRules.EntityName.Label)
		if err != nil {
			return err
		}
		e, ok := ebn[serviceName]
		if !ok {
			continue
		}

		attributes, metadata := getAttributesAndMetadata(entityRules, metricRules.Attributes, m, hostname)
		addMetadata(metadata, e)
		// _info metrics only contains metadata
		if metricRules.InfoMetric {
			continue
		}

		metricName := metricRules.NrdbName
		gauge, err := integration.Gauge(time.Now(), metricName, metricValue)
		warnOnErr(err)
		addAttributes(attributes, gauge)
		e.AddMetric(gauge)
		noMetricAdded = false
	}
	if noMetricAdded && metricRules.EnumMetric {
		log.Debug("all metrics have value 0 for: %s", metricFamily.GetName())
	}
	return nil
}

func addMetadata(metadata metadataMap, e *integration.Entity) {
	var err error
	for k, v := range metadata {
		err = e.AddMetadata(k, v)
		warnOnErr(err)
	}
}
func addAttributes(attributes attributesMap, metric metric.Metric) {
	var err error
	for k, v := range attributes {
		err = metric.AddDimension(k, v)
		warnOnErr(err)
	}
}

func getAttributesAndMetadata(entityRules EntityRules, attributesRules []Attribute, metric *dto.Metric, hostname string) (attributesMap, metadataMap) {
	var metadata = make(map[string]string)
	var attributes = make(map[string]string)

	metadata[entityRules.EntityName.HostnameNrdbLabelName] = hostname

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

func warnOnErr(err error) {
	if err != nil {
		log.Warn(err.Error())
	}
}

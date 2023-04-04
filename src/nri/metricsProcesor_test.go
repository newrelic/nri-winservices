/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package nri

import (
	"testing"

	"github.com/newrelic/infra-integrations-sdk/v4/integration"
	"github.com/newrelic/nri-winservices/src/matcher"
	"github.com/newrelic/nri-winservices/src/scraper"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	serviceName        = "rpcss"
	serviceStartMode   = "auto"
	serviceDisplayName = "Remote Procedure Call (RPC)"
	servicePid         = "668"
	hostname           = "test-hostname"
)

var filter = []string{serviceName}
var gauge = dto.MetricType_GAUGE

var metricFamlilyServiceInfo = dto.MetricFamily{
	Name: strPtr("windows_service_info"),
	Type: &gauge,
	Metric: []*dto.Metric{
		{
			Label: []*dto.LabelPair{
				{
					Name:  strPtr("name"),
					Value: strPtr(serviceName),
				},
				{
					Name:  strPtr("display_name"),
					Value: strPtr(serviceDisplayName),
				},
				{
					Name:  strPtr("process_id"),
					Value: strPtr(servicePid),
				},
			},
			Gauge: &dto.Gauge{
				Value: float64Ptr(1),
			},
		},
	},
}
var metricFamlilyService = dto.MetricFamily{
	Name: strPtr("windows_service_start_mode"),
	Type: &gauge,
	Metric: []*dto.Metric{
		{
			Label: []*dto.LabelPair{
				{
					Name:  strPtr("name"),
					Value: strPtr(serviceName),
				},
				{
					Name:  strPtr("start_mode"),
					Value: strPtr(serviceStartMode),
				},
			},
			Gauge: &dto.Gauge{
				Value: float64Ptr(1),
			},
		},
		{
			Label: []*dto.LabelPair{
				{
					Name:  strPtr("name"),
					Value: strPtr("rpcss"),
				},
				{
					Name:  strPtr("start_mode"),
					Value: strPtr("boot"),
				},
			},
			Gauge: &dto.Gauge{
				Value: float64Ptr(0),
			},
		},
	},
}

func TestCreateEntities(t *testing.T) {
	i, _ := integration.New("integrationName", "integrationVersion")
	rules := loadRules()
	mfbn := scraper.MetricFamiliesByName{
		"windows_service_info":       metricFamlilyServiceInfo,
		"windows_service_start_mode": metricFamlilyService,
	}

	matcher := matcher.New(filter)
	entityMap, err := createEntities(i, mfbn, rules, matcher)
	require.NoError(t, err)
	_, ok := entityMap[serviceName]
	require.True(t, ok)
	require.Len(t, i.Entities, 1)
	require.Equal(t, i.Entities[0].Name(), entityNamePrefix+":"+hostName+":"+serviceName)
	require.False(t, i.Entities[0].IgnoreEntity)
}

func TestNoServiceNameAllowed(t *testing.T) {
	i, _ := integration.New("integrationName", "integrationVersion")
	rules := loadRules()
	mfbn := scraper.MetricFamiliesByName{
		"windows_service_info":       metricFamlilyServiceInfo,
		"windows_service_start_mode": metricFamlilyService,
	}

	matcher := matcher.New([]string{})
	entityMap, err := createEntities(i, mfbn, rules, matcher)
	require.NoError(t, err, "No error is expected even if no service is allowed")
	require.Len(t, entityMap, 0, "No entity is expected since no service is allowed")
	err = processMetricGauge(metricFamlilyService, rules, entityMap, hostname)
	err = processMetricGauge(metricFamlilyService, rules, entityMap, hostname)
	require.NoError(t, err)
	require.NoError(t, err, "No error is expected even if entityMap is empty")
}

func TestProccessMetricGauge(t *testing.T) {
	i, _ := integration.New("integrationName", "integrationVersion")
	rules := loadRules()
	mfbn := scraper.MetricFamiliesByName{
		"windows_service_info":       metricFamlilyServiceInfo,
		"windows_service_start_mode": metricFamlilyService,
	}

	matcher := matcher.New(filter)
	entityMap, err := createEntities(i, mfbn, rules, matcher)
	require.NoError(t, err)
	// process info metrics
	err = processMetricGauge(metricFamlilyServiceInfo, rules, entityMap, hostname)
	require.NoError(t, err)
	metadata := entityMap[serviceName].GetMetadata()
	assert.Equal(t, serviceDisplayName, metadata["display_name"])
	assert.Equal(t, servicePid, metadata["process_id"])
	// process startmode metrics
	err = processMetricGauge(metricFamlilyService, rules, entityMap, hostname)
	assert.NoError(t, err)
	assert.Equal(t, serviceStartMode, metadata["start_mode"])

}

func strPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

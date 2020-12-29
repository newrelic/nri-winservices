/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package nri

import (
	"testing"
	"time"

	"github.com/newrelic/infra-integrations-sdk/v4/integration"
	"github.com/newrelic/nri-winservices/src/matcher"
	"github.com/newrelic/nri-winservices/src/scraper"
	"github.com/stretchr/testify/require"
)

func TestProccessInventory(t *testing.T) {
	entityRules := loadRules()
	i, _ := integration.New("integrationName", "integrationVersion")
	mfbn := scraper.MetricFamiliesByName{
		"windows_service_info":       metricFamlilyServiceInfo,
		"windows_service_start_mode": metricFamlilyService,
		"windows_cs_hostname":        metricFamlilyServiceHostname,
	}

	matcher := matcher.New(filter)
	err := ProcessMetrics(i, mfbn, matcher)
	require.NoError(t, err)
	require.Greater(t, len(i.Entities), 0)

	err = ProcessInventory(i)
	require.NoError(t, err)
	require.Greater(t, len(i.Entities), 0)

	require.Equal(t, hostname+":"+serviceName, i.Entities[0].Name())

	item, ok := i.Entities[0].Inventory.Item(entityTypeInventory)
	require.True(t, ok)
	require.Equal(t, hostname, item[entityRules.EntityName.HostnameNrdbLabelName])
	require.Equal(t, hostname+":"+serviceName, item["name"])
	require.Equal(t, serviceName, item["windowsService.name"])
	require.Equal(t, serviceDisplayName, item["windowsService.displayName"])
	require.Equal(t, servicePid, item["windowsService.processId"])
	require.Equal(t, serviceStartMode, item["windowsService.startMode"])
	require.Equal(t, time.Now().Hour(), item[heartBeatInventory])

}

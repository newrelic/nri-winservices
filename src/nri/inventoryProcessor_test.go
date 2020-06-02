package nri

import (
	"testing"

	"github.com/newrelic/infra-integrations-sdk/integration"
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

	validator := NewValidator(serviceName, "", "")
	err := ProcessMetrics(i, mfbn, validator)
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

}

/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package scraper

import (
	"strconv"

	dto "github.com/prometheus/client_model/go"
	"golang.org/x/sys/windows/svc/mgr"
)

var (
	windowsServiceInfo      = "windows_service_info"
	windowsServiceStartMode = "windows_service_start_mode"
	windowsServiceState     = "windows_service_state"
	gauge                   = dto.MetricType_GAUGE
	gaugeValue              = float64(1)
	labelServiceName        = "name"
	labelDisplayName        = "display_name"
	labelProcessID          = "process_id"
	labelRunAs              = "run_as"
	labelState              = "state"
	stateDescription        = []string{"pending", "pending", "paused", "running", "starting", "stopping", "stopped"}
)

func GetServices() (MetricFamiliesByName, error) {
	// Open connection to Services Manager
	svcConnection, err := mgr.Connect()
	if err != nil {
		return nil, err
	}
	defer svcConnection.Disconnect()

	// List All Services from the Services Manager
	svcList, err := svcConnection.ListServices()
	if err != nil {
		return nil, err
	}

	mfs := MetricFamiliesByName{}
	metrics := []*dto.Metric{}

	// Iterate through the Services List
	for _, svc := range svcList {
		// Retrieve the handle for each service
		svcHandle, err := svcConnection.OpenService(svc)
		if err != nil {
			continue
		}

		// Get Service Configuration
		svcConfig, err := svcHandle.Config()
		if err != nil {
			_ = svcHandle.Close()
			continue
		}

		// Get Service Current Status
		svcStatus, err := svcHandle.Query()
		if err != nil {
			_ = svcHandle.Close()
			continue
		}

		name := svc
		displayName := svcConfig.DisplayName
		runAs := svcConfig.ServiceStartName
		servicePid := strconv.FormatUint(uint64(svcStatus.ProcessId), 10)
		// state := stateDescription[svcStatus.State]

		metrics = append(metrics, &dto.Metric{
			Label: []*dto.LabelPair{
				{
					Name:  &labelDisplayName,
					Value: &displayName,
				},
				{
					Name:  &labelServiceName,
					Value: &name,
				},
				{
					Name:  &labelProcessID,
					Value: &servicePid,
				},
				{
					Name:  &labelRunAs,
					Value: &runAs,
				},
			},
			Gauge: &dto.Gauge{
				Value: &gaugeValue,
			},
			TimestampMs: nil,
		})

		_ = svcHandle.Close()
	}

	//windows_service_info{"display_name"="IP Helper",name="iphlpsvc",process_id="2340",run_as="LocalSystem"} 1
	mfs[windowsServiceInfo] = dto.MetricFamily{
		Name:   &windowsServiceInfo,
		Type:   &gauge,
		Metric: metrics,
	}
	return mfs, nil
}

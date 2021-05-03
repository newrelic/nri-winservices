/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package scraper

import (
	"strconv"

	dto "github.com/prometheus/client_model/go"
	"golang.org/x/sys/windows"
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
	labelStartMode          = "start_mode"
	labelStartModeValues    = map[uint32]string{
		windows.SERVICE_AUTO_START:   "auto",
		windows.SERVICE_BOOT_START:   "boot",
		windows.SERVICE_DEMAND_START: "manual",
		windows.SERVICE_DISABLED:     "disabled",
		windows.SERVICE_SYSTEM_START: "system",
	}
	labelState       = "state"
	labelStateValues = map[uint]string{
		windows.SERVICE_CONTINUE_PENDING: "pending",
		windows.SERVICE_PAUSE_PENDING:    "pending",
		windows.SERVICE_PAUSED:           "paused",
		windows.SERVICE_RUNNING:          "running",
		windows.SERVICE_START_PENDING:    "starting",
		windows.SERVICE_STOP_PENDING:     "stopping",
		windows.SERVICE_STOPPED:          "stopped",
	}
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
	metricsInfo := []*dto.Metric{}
	metricsStartMode := []*dto.Metric{}
	metricsState := []*dto.Metric{}

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
		pid := strconv.FormatUint(uint64(svcStatus.ProcessId), 10)
		startMode := labelStartModeValues[svcConfig.StartType]
		state := labelStateValues[uint(svcStatus.State)]

		metricsInfo = append(metricsInfo, &dto.Metric{
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
					Value: &pid,
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

		metricsStartMode = append(metricsStartMode, &dto.Metric{
			Label: []*dto.LabelPair{
				{
					Name:  &labelServiceName,
					Value: &name,
				},
				{
					Name:  &labelStartMode,
					Value: &startMode,
				},
			},
			Gauge: &dto.Gauge{
				Value: &gaugeValue,
			},
			TimestampMs: nil,
		})

		metricsState = append(metricsState, &dto.Metric{
			Label: []*dto.LabelPair{
				{
					Name:  &labelServiceName,
					Value: &name,
				},
				{
					Name:  &labelState,
					Value: &state,
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
		Metric: metricsInfo,
	}

	//windows_service_start_mode{name="aarsvc_390e7",start_mode="auto"} 0
	mfs[windowsServiceStartMode] = dto.MetricFamily{
		Name:   &windowsServiceStartMode,
		Type:   &gauge,
		Metric: metricsStartMode,
	}

	//windows_service_state{name="aarsvc_390e7",state="continue pending"} 0
	mfs[windowsServiceState] = dto.MetricFamily{
		Name:   &windowsServiceState,
		Type:   &gauge,
		Metric: metricsState,
	}

	return mfs, nil
}

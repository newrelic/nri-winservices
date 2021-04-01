/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package scraper

import (
	dto "github.com/prometheus/client_model/go"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	key    = registry.LOCAL_MACHINE
	path   = `SYSTEM\CurrentControlSet\Services\`
	access = registry.QUERY_VALUE | registry.ENUMERATE_SUB_KEYS

	// defined in wdm.h (winsdk-10)
	//SERVICE_USER_SERVICE = 0x00000040
	//SERVICE_USERSERVICE_INSTANCE = 0x00000080
)

var (
	//muiErr = registry.LoadRegLoadMUIString()
	windowsServiceInfo = "windows_service_info"
	gauge              = dto.MetricType_GAUGE
	gaugeValue         = float64(1)
	labelServiceName   = "name"
	labelDisplayName   = "display_name"
	labelProcessID     = "process_id"
	labelRunAs         = "run_as"
	servicePid         = "0"
)

type service struct {
	name        string
	displayName string
}

func GetRegistry() (MetricFamiliesByName, error) {
	k, err := registry.OpenKey(key, path, access)
	if err != nil {
		return nil, err
	}
	defer func() { _ = k.Close() }()

	services, err := k.ReadSubKeyNames(0)
	if err != nil {
		return nil, err
	}

	mfs := MetricFamiliesByName{}
	metrics := []*dto.Metric{}
	for _, name := range services {
		name := name

		svc, err := registry.OpenKey(key, path+name, access)
		if err != nil {
			continue
		}

		t, _, err := svc.GetIntegerValue("Type")
		if t&windows.SERVICE_WIN32 == 0 || err != nil {
			_ = svc.Close()
			continue
		}

		//registry.LoadRegLoadMUIString()
		displayName, err := svc.GetMUIStringValue("DisplayName")
		if err != nil {
			displayName, _, err = svc.GetStringValue("DisplayName")
			if err != nil {
				_ = svc.Close()
				continue
			}
		}

		// ObjectName is a type REG_DWORD which contains the account name for services or the driver object that the I/O manager uses to load the device driver.
		runAs, _, err := svc.GetStringValue("ObjectName")
		if err != nil {
			_ = svc.Close()
			continue
		}

		//fmt.Println(displayName, runAs)
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

		_ = svc.Close()
	}

	//windows_service_info{"display_name"="IP Helper",name="iphlpsvc",process_id="2340",run_as="LocalSystem"} 1
	mfs[windowsServiceInfo] = dto.MetricFamily{
		Name:   &windowsServiceInfo,
		Type:   &gauge,
		Metric: metrics,
	}

	return mfs, nil
}

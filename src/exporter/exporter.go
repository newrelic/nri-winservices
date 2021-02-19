/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/newrelic/infra-integrations-sdk/v4/log"
	"golang.org/x/sys/windows"
)

const (
	// ExporterName name of the exporter binary
	ExporterName      = "windows_exporter.exe"
	enabledCollectors = "service"
	logFormat         = "exporter msg=%v source=%v"
)

// Exporter manages the exporter execution
type Exporter struct {
	URL        string
	MetricPath string
	Done       chan struct{} // this channel is closed when the exporter stop running
	cmd        *exec.Cmd
	ctx        context.Context
	cancel     context.CancelFunc
	jobObject  windows.Handle
}

type process struct {
	pid    int
	handle windows.Handle
}

// New create a configured Exporter struct ready to be run
func New(verbose bool, bindAddress string, bindPort string) (*Exporter, error) {

	integrationDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter:%v", err)
	}
	exporterPath := filepath.Join(integrationDir, ExporterName)

	ctx, cancel := context.WithCancel(context.Background())

	exporterLogLevel := "info"
	if verbose {
		exporterLogLevel = "debug"
	}
	exporterURL := bindAddress + ":" + bindPort

	cmd := exec.CommandContext(ctx,
		exporterPath,
		"--collectors.enabled", enabledCollectors,
		"--log.level", exporterLogLevel,
		"--log.format", "logger:stderr?json=true",
		"--collector.service.services-where", "Name like '%'", //All Added to avoid warn message from Exporter
		"--telemetry.addr", exporterURL)

	return &Exporter{
		URL:        exporterURL,
		MetricPath: "/metrics",
		cmd:        cmd,
		ctx:        ctx,
		cancel:     cancel,
		Done:       make(chan struct{}),
	}, nil
}

// Run executes the exporter binary
func (e *Exporter) Run() error {
	e.redirectLogs()
	err := e.cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to run exporter:%v", err)
	}
	if err = e.createJobObject(); err != nil {
		return fmt.Errorf("failed to create job object:%v", err)
	}
	go func() {
		e.cmd.Wait()
		log.Debug("exporter has stopped")
		close(e.Done)
	}()

	return nil
}

// createJobObject adds the process to a JobObject configured to kill the process
// when the parent is killed
func (e *Exporter) createJobObject() error {
	var err error
	e.jobObject, err = windows.CreateJobObject(nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create job object: %v", err)
	}

	jobInfo := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
		},
	}
	_, err = windows.SetInformationJobObject(
		e.jobObject,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&jobInfo)),
		uint32(unsafe.Sizeof(jobInfo)),
	)
	if err != nil {
		return fmt.Errorf("failed to set job object info:%v", err)
	}

	err = windows.AssignProcessToJobObject(e.jobObject, (*process)(unsafe.Pointer(e.cmd.Process)).handle)
	if err != nil {
		return fmt.Errorf("failed to assign process to job object:%v", err)
	}

	return nil
}

// Kill cancel the ctx and close the handle
func (e *Exporter) Kill() {
	windows.CloseHandle(e.jobObject)
	select {
	case <-e.Done: //exporter is not running any more
		return
	default:
		e.cancel()
	}
}

type logMsg struct {
	Level, Msg, Source string
}

func (e *Exporter) redirectLogs() {
	r, err := e.cmd.StderrPipe()
	if err != nil {
		log.Error(err.Error())
	}
	var m logMsg
	dec := json.NewDecoder(r)
	go func() {
		for {
			if err := dec.Decode(&m); err == io.EOF {
				log.Debug("cmd StderrPipe has closed")
				return
			} else if err != nil {
				log.Error("failed decoding exporter log:%v", err)
			}
			switch m.Level {
			case "debug":
				log.Debug(logFormat, m.Msg, m.Source)
			case "info":
				log.Info(logFormat, m.Msg, m.Source)
			case "warn":
				log.Warn(logFormat, m.Msg, m.Source)
			case "error":
				// TODO currently the exporter detects that is being lunched from a non interactive session (by the Agent)
				// and tries to register as a service but fails. This is not affecting the exporter nither leaving Windows Event logs
				// This should be removed after modify the exporter behavior when is lunched from other process.
				if strings.Contains(m.Msg, "Failed to start service: The service process could not connect to the service controller") {
					// we remove this log since it can be misleading.
					continue
				}
				log.Error(logFormat, m.Msg, m.Source)
			case "fatal":
				// on Fatal error cmd.Wait() ends closing Done channel
				msg := m.Msg + "(fatal error on exporter)"
				log.Error(logFormat, msg, m.Source)
			}
		}
	}()
}

package exporter

import (
	"bufio"
	"context"
	"os/exec"
	"strings"
	"unsafe"

	"github.com/newrelic/infra-integrations-sdk/log"
	"golang.org/x/sys/windows"
)

const (
	//TODO update this with the corresponding path
	exporterPath      = "C:\\Program Files\\New Relic\\newrelic-infra\\newrelic-integrations\\bin\\wmi_exporter.exe"
	enabledCollectors = "service,cs"
)

// Exporter manages the exporter execution
type Exporter struct {
	URL        string
	MetricPath string
	cmd        *exec.Cmd
	ctx        context.Context
	cancel     context.CancelFunc
	jobObject  windows.Handle
	Done       chan struct{} // this channel is closed when the exporter stop running
}

type process struct {
	pid    int
	handle windows.Handle
}

// New create a configured Exporter struct ready to be runned
func New(verbose bool, bindAddress string, bindPort string) Exporter {
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
		"--collector.service.services-where", "Name like '%'", //All Added to avoid warn message from Exporter
		"--telemetry.addr", exporterURL)

	r, err := cmd.StderrPipe()
	if err != nil {
		log.Error(err.Error())
	}
	reader := bufio.NewReader(r)
	go func() {
		for {
			exporterLog, err := reader.ReadString('\n')
			if err != nil {
				return // terminate on EOF
			}
			// TODO currently the exporter detects that is being lunched from a non interactive session (by the Agent)
			// and tries to register as a service but fails. This is not affecting the exporter nither leaving Windows Event logs
			// This should remove after modify the exporter behavior when is lunched from other process.
			if strings.Contains(exporterLog, "Failed to start service: The service process could not connect to the service controller") {
				// we remove this log since could be misslead a wrong interpretation.
				continue
			}
			log.Info(exporterLog)
		}
	}()
	return Exporter{
		URL:        exporterURL,
		MetricPath: "/metrics",
		cmd:        cmd,
		ctx:        ctx,
		cancel:     cancel,
		Done:       make(chan struct{}),
	}
}

// Run executes the exporter binary
func (e *Exporter) Run() error {
	err := e.cmd.Start()
	if err != nil {
		return err
	}
	if err := e.createJobObject(); err != nil {
		return err
	}
	go func() {
		e.cmd.Wait()
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
		return err
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
		return err
	}

	windows.AssignProcessToJobObject(e.jobObject, (*process)(unsafe.Pointer(e.cmd.Process)).handle)
	if err != nil {
		return err
	}

	return nil
}

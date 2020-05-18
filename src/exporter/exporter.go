package exporter

import (
	"bufio"
	"context"
	"os/exec"
	"unsafe"

	"github.com/newrelic/infra-integrations-sdk/log"
	"golang.org/x/sys/windows"
)

const (
	exporterPath      = "wmi_exporter.exe"
	enabledCollectors = "service,cs"
)

// Exporter manages the exporter execution
type Exporter struct {
	cmd       *exec.Cmd
	ctx       context.Context
	cancel    context.CancelFunc
	jobObject windows.Handle
}

type process struct {
	pid    int
	handle windows.Handle
}

// New create a configured Exporter struct ready to be runned
func New(verbose bool, bindingAddress string) Exporter {
	ctx, cancel := context.WithCancel(context.Background())

	exporterLogLevel := "info"
	if verbose {
		exporterLogLevel = "debug"
	}

	cmd := exec.CommandContext(ctx,
		exporterPath,
		"--collectors.enabled", enabledCollectors,
		"--log.level", exporterLogLevel,
		"--collector.service.services-where", "Name like '%'", //All Added to avoid warn message from Exporter
		"--telemetry.addr", bindingAddress)

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
			log.Info(exporterLog)
		}
	}()
	return Exporter{cmd: cmd, ctx: ctx, cancel: cancel}
}

// Run executes the exporter binary
func (e *Exporter) Run() {
	err := e.cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	if err := e.createJobObject(); err != nil {
		log.Error(err.Error())
	}
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

// Kill cancel the ctx and close the handle
func (e *Exporter) Kill() {
	windows.CloseHandle(e.jobObject)
	e.cancel()
	if err := e.cmd.Wait(); err != nil {
		log.Error(err.Error())
	}
}

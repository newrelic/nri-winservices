// +build integration

package test

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/newrelic/nri-winservices/src/exporter"
	"github.com/newrelic/nri-winservices/test/jsonschema"
	"github.com/stretchr/testify/assert"
)

func isProcessRunning(name string) bool {
	out, err := exec.Command("tasklist").Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.Contains(string(out), name)
}

func runIntegration() (string, string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(
		ctx,
		"./nri-winservices.exe",
		"-scrape_interval", "15s",
		"-exporter_bind_address", "127.0.0.1",
		"-exporter_bind_port", "9183",
		// "-allow_regex", "dmwappushservice",
		"-allow_regex", "^*$",
	)
	defer cmd.Wait()
	defer cancel()

	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Start()
	if err != nil {
		return "", "", fmt.Errorf("fail to start cmd: %v", err)
	}

	time.Sleep(17 * time.Second)
	stdout := outbuf.String()
	stderr := errbuf.String()

	stdout = strings.ReplaceAll(stdout, "{}\n", "")
	return stdout, stderr, nil

	// timeout := time.NewTicker(17 * time.Second)
	// for {
	// 	select {
	// 	case <-timeout.C:
	// 		return "", "", fmt.Errorf("fail to execute integration: timeout reached")
	// 	default:
	// 		l, err := outbuf.ReadString('\n')
	// 		if err != nil {
	// 			if err == io.EOF {
	// 				continue
	// 			}
	// 			return "", "", fmt.Errorf("fail to read out: %v", err)
	// 		}
	// 		if l == "{}" {
	// 			continue
	// 		}
	// 		stdout := outbuf.String()
	// 		stderr := errbuf.String()

	// 		// stdout = strings.ReplaceAll(stdout, "{}\n", "")
	// 		return stdout, stderr, nil
	// 	}
	// }
}
func TestIntegration(t *testing.T) {
	assert.False(t, isProcessRunning(exporter.ExporterName))

	stdout, stderr, err := runIntegration()
	assert.NotNil(t, stderr, "unexpected stderr")
	assert.NoError(t, err, "Unexpected error")

	schemaPath := filepath.Join("json-schema-files", "winservices-schema.json")

	fmt.Print(stdout)
	err = jsonschema.Validate(schemaPath, stdout)
	assert.NoError(t, err, "The output of WinServices integration doesn't have expected format.")

	assert.False(t, isProcessRunning(exporter.ExporterName))
}

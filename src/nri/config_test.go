package nri

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	content := []byte(`
exporter_bind_address: 127.0.0.1
exporter_bind_port: 9182
scrape_interval: 30s
filter_entity:
  windowsService.name:
    - regex ".*"
    - "ServiceNameToBeIncluded"
    - not "ServiceNameToBeExcluded"`)

	tmpfile, err := ioutil.TempFile("", "config")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name()) // clean up
	_, err = tmpfile.Write(content)
	require.NoError(t, err)

	_, err = NewConfig(tmpfile.Name())
	require.NoError(t, err)
}

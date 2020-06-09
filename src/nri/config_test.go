package nri

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseConfigYaml(t *testing.T) {
	content := []byte(`
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

	c, err := ParseConfigYaml(tmpfile.Name())
	require.NoError(t, err)
	fmt.Printf("%v", c)

}

// func TestNew(t *testing.T) {
// 	yml := `
// filter_entity:
//   windowsService.name:
//     - regex ".*"
//     - "ServiceNameToBeIncluded"
//     - not "ServiceNameToBeExcluded"`
// 	New(yml)
// }

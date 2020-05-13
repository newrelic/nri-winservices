package scraper

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetReal(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/actualOutput")
	}))
	defer ts.Close()
	mfs, err := Get(http.DefaultClient, ts.URL)
	var actual []string
	for k := range mfs {
		actual = append(actual, k)
	}
	assert.NoError(t, err)
	fmt.Println(*mfs["wmi_os_processes"].Name)
	fmt.Println(mfs["wmi_os_processes"].Metric[0])
	fmt.Println(*mfs["wmi_os_processes"].Help)
	fmt.Println("EXAMPLE")
}

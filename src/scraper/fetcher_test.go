package scraper

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/simple-metrics")
	}))
	defer ts.Close()
	expected := []string{"go_goroutines", "go_memstats_heap_idle_bytes", "go_gc_duration_seconds", "http_requests_total"}
	mfs, err := Get(http.DefaultClient, ts.URL)
	var actual []string
	for k := range mfs {
		actual = append(actual, k)
	}
	assert.NoError(t, err)
	assert.ElementsMatch(t, expected, actual)
}

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

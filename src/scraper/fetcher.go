package scraper

import (
	dto "github.com/prometheus/client_model/go"
	"io"
	"net/http"

	"github.com/prometheus/common/expfmt"
)

type MetricFamiliesByName map[string]dto.MetricFamily

// Get scrapes the given URL and decodes the retrieved payload.
func Get(client HTTPDoer, url string) (MetricFamiliesByName, error) {
	mfs := MetricFamiliesByName{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return mfs, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return mfs, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	countedBody := &countReadCloser{innerReadCloser: resp.Body}
	d := expfmt.NewDecoder(countedBody, expfmt.FmtText)
	for {
		var mf dto.MetricFamily
		if err := d.Decode(&mf); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		mfs[mf.GetName()] = mf
	}
	return mfs, nil
}

// HTTPDoer executes http requests. It is implemented by *http.Client.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type countReadCloser struct {
	innerReadCloser io.ReadCloser
	count           int
}

func (rc *countReadCloser) Close() error {
	return rc.innerReadCloser.Close()
}

func (rc *countReadCloser) Read(p []byte) (n int, err error) {
	n, err = rc.innerReadCloser.Read(p)
	rc.count += n
	return
}

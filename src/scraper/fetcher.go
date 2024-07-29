/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package scraper

import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v4/log"
	dto "github.com/prometheus/client_model/go"
	"io"
	"net/http"
	"time"

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
	log.Debug("Performing HTTP request against: %s", url)
	t := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return mfs, err
	}
	if resp.StatusCode != 200 {
		return mfs, fmt.Errorf("the exporter answered with a value different from 200")
	}
	log.Debug("HTTP request performed - Status: %s, total time taken to perform request: %s", resp.Status, time.Since(t).String())

	defer func() {
		_ = resp.Body.Close()
	}()

	log.Debug("Parsing body of the exporter answer")
	countedBody := &countReadCloser{innerReadCloser: resp.Body}
	d := expfmt.NewDecoder(countedBody, expfmt.NewFormat(expfmt.TypeTextPlain))
	for {
		var mf dto.MetricFamily
		if err = d.Decode(&mf); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		mfs[mf.GetName()] = mf
	}
	log.Debug("Body of the exporter answer parsed")
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

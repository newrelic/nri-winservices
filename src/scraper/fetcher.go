package scraper

import (
	dto "github.com/prometheus/client_model/go"
	prometheusclient "github.com/prometheus/client_model/go"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

type metricValue interface{}
type metricType string

//nolint:golint
const (
	metricType_COUNTER   metricType = "count"
	metricType_GAUGE     metricType = "gauge"
	metricType_SUMMARY   metricType = "summary"
	metricType_HISTOGRAM metricType = "histogram"
)

// Metric represents a Prometheus metric.
// https://prometheus.io/docs/concepts/data_model/
type Metric struct {
	name       string
	value      metricValue
	metricType metricType
	attributes Set
}

var supportedMetricTypes = map[prometheusclient.MetricType]string{
	prometheusclient.MetricType_COUNTER:   "counter",
	prometheusclient.MetricType_GAUGE:     "gauge",
	prometheusclient.MetricType_HISTOGRAM: "histogram",
	prometheusclient.MetricType_SUMMARY:   "summary",
	prometheusclient.MetricType_UNTYPED:   "untyped",
}

type MetricFamiliesByName map[string]dto.MetricFamily

func convertPromMetrics(log *logrus.Entry, targetName string, mfs MetricFamiliesByName) []Metric {
	var metricsCap int
	for _, mf := range mfs {
		mtype, ok := supportedMetricTypes[mf.GetType()]
		if !ok {
			continue
		}
		metricsCap += len(mf.Metric)
		totalTimeseriesByTargetAndTypeMetric.WithLabelValues(mtype, targetName).Add(float64(len(mf.Metric)))
		totalTimeseriesByTypeMetric.WithLabelValues(mtype).Add(float64(len(mf.Metric)))
		totalTimeseriesByTargetMetric.WithLabelValues(targetName).Add(float64(len(mf.Metric)))
	}
	totalTimeseriesMetric.Add(float64(metricsCap))

	metrics := make([]Metric, 0, metricsCap)
	for mname, mf := range mfs {
		ntype := mf.GetType()
		mtype, ok := supportedMetricTypes[ntype]
		if !ok {
			continue
		}
		for _, m := range mf.GetMetric() {
			var value interface{}
			var nrType metricType
			switch ntype {
			case prometheusclient.MetricType_UNTYPED:
				value = m.GetUntyped().GetValue()
				nrType = metricType_GAUGE
			case prometheusclient.MetricType_COUNTER:
				value = m.GetCounter().GetValue()
				nrType = metricType_COUNTER
			case prometheusclient.MetricType_GAUGE:
				value = m.GetGauge().GetValue()
				nrType = metricType_GAUGE
			case prometheusclient.MetricType_SUMMARY:
				value = m.GetSummary()
				nrType = metricType_SUMMARY
			case prometheusclient.MetricType_HISTOGRAM:
				value = m.GetHistogram()
				nrType = metricType_HISTOGRAM
			default:
				if log.Level <= logrus.DebugLevel {
					log.WithField("target", targetName).Debugf("metric type not supported: %s", mtype)
				}
				continue
			}
			attrs := map[string]interface{}{}
			attrs["targetName"] = targetName
			for _, l := range m.GetLabel() {
				attrs[l.GetName()] = l.GetValue()
			}
			attrs["nrMetricType"] = string(nrType)
			attrs["promMetricType"] = mtype
			metrics = append(
				metrics,
				Metric{
					name:       mname,
					metricType: nrType,
					value:      value,
					attributes: attrs,
				},
			)
		}
	}
	return metrics
}

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

	bodySize := float64(countedBody.count)
	targetSize.With(prom.Labels{"target": url}).Set(bodySize)
	totalScrapedPayload.Add(bodySize)
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

// ResetTotalScrapedPayload resets the integration totalScrapedPayload
// metric.
func ResetTotalScrapedPayload() {
	totalScrapedPayload.Set(0)
}

var (
	targetSize = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: "nr_stats",
		Subsystem: "integration",
		Name:      "payload_size",
		Help:      "Size of target's payload",
	},
		[]string{
			"target",
		},
	)
	totalScrapedPayload = prom.NewGauge(prom.GaugeOpts{
		Namespace: "nr_stats",
		Subsystem: "integration",
		Name:      "total_payload_size",
		Help:      "Total size of the payloads scraped",
	})
)

func init() {
	prom.MustRegister(targetSize)
	prom.MustRegister(totalScrapedPayload)
}

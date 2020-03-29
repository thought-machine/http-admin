package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_model/go"
)

// Gatherer is the thing we gather metrics from.
var Gatherer = prometheus.DefaultGatherer

func renderMetrics(w http.ResponseWriter, keys sort.StringSlice) {
	content := `<link type="text/css" href="/admin/files/css/metric-query.css" rel="stylesheet"/>
        <script type="application/javascript" src="/admin/files/js/metric-query.js"></script>
        <script type="application/javascript" src="/admin/files/js/chart-renderer.js"></script>
        <div id="metrics-grid" class="row" data-refresh-uri="/admin/metrics">
          <div class="col-md-4 snuggle-right">
            <ul id="metrics" class="list-unstyled">`
	sort.Sort(keys)
	for _, key := range keys {
		content += fmt.Sprintf(`<li id="%s">%s</li>`+"\n", strings.Replace(key, "/", "-", -1), key)
	}
	content += `</ul>
          </div>
          <div class="col-md-8 snuggle-left">
            <div id="chart-div"></div>
          </div>
        </div>`

	w.Write([]byte(content))
}

type statEntry struct {
	Name  string   `json:"name"`
	Value *float64 `json:"value"`
}

func query(mfs []*io_prometheus_client.MetricFamily, ms map[string]struct{}) []statEntry {
	ret := []statEntry{}
	for _, mf := range mfs {
		if _, present := ms[mf.GetName()]; present {
			metrics := mf.GetMetric()
			for _, m := range metrics {
				seriesName := mf.GetName()
				if len(m.Label) != 0 {
					labelValues := make([]string, 0)
					for _, v := range m.Label {
						labelValues = append(labelValues, v.GetName()+"="+v.GetValue())
					}
					seriesName += "{" + strings.Join(labelValues, ",") + "}"
				}
				switch mf.GetType() {
				case io_prometheus_client.MetricType_COUNTER:
					ret = append(ret, statEntry{Name: seriesName, Value: m.Counter.Value})
				case io_prometheus_client.MetricType_GAUGE:
					ret = append(ret, statEntry{Name: seriesName, Value: m.Gauge.Value})
				case io_prometheus_client.MetricType_SUMMARY:
					ret = append(ret, statEntry{Name: seriesName, Value: m.Summary.SampleSum})
				case io_prometheus_client.MetricType_HISTOGRAM:
					ret = append(ret, statEntry{Name: seriesName, Value: m.Histogram.SampleSum})
				}
			}
		}
	}
	return ret
}

// MetricQueryHandler either renders the list of all metrics and a graph, or returns the queried metrics' current values.
func MetricQueryHandler(w http.ResponseWriter, r *http.Request) {
	mfs, err := Gatherer.Gather()
	ms, present := r.URL.Query()["m"]
	if !present {
		keys := []string{}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, mf := range mfs {
			keys = append(keys, *mf.Name)
		}
		writeContentType(w, "text/html;charset=UTF-8")
		renderMetrics(w, keys)
	} else {
		writeContentType(w, "application/json;charset=UTF-8")
		mMap := make(map[string]struct{})
		for _, m := range ms {
			mMap[m] = struct{}{}
		}
		b, _ := json.Marshal(query(mfs, mMap))
		w.Write(b)
	}
}

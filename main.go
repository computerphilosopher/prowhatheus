package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/common/model"

	"github.com/prometheus/prometheus/storage/remote"
)

const (
	license = "x4pgn21h11dnu-x235qldgcfs9iv-z37dl208p063vq"
)

func handler(w http.ResponseWriter, r *http.Request) {
	req, err := remote.DecodeWriteRequest(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, ts := range req.Timeseries {
		m := make(model.Metric, len(ts.Labels))
		for _, l := range ts.Labels {
			m[model.LabelName(l.Name)] = model.LabelValue(l.Value)
		}
		fmt.Println(m)

		for _, s := range ts.Samples {
			fmt.Printf("\tSample:  %f %d\n", s.Value, s.Timestamp)
		}

		for _, e := range ts.Exemplars {
			m := make(model.Metric, len(e.Labels))
			for _, l := range e.Labels {
				m[model.LabelName(l.Name)] = model.LabelValue(l.Value)
			}
			fmt.Printf("\tExemplar:  %+v %f %d\n", m, e.Value, e.Timestamp)
		}

		for _, hp := range ts.Histograms {
			h := remote.HistogramProtoToHistogram(hp)
			fmt.Printf("\tHistogram:  %s\n", h.String())
		}
	}

}

func main() {
	http.HandleFunc("/receive", handler)
	log.Fatal("listen 0.0.0.0:19090", http.ListenAndServe("0.0.0.0:19090", nil))
}

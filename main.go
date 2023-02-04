package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/whatap/golib/net/oneway"

	"github.com/whatap/golib/lang/pack"
	"github.com/whatap/golib/logger"
	whash "github.com/whatap/golib/util/hash"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
)

var (
	license    = ""
	pcode      = int64(0)
	oname      = ""
	whatapHost = ""
	listen     = ""
)

var tcpClient *oneway.OneWayTcpClient

func genPackTemplates(labels []prompb.Label, samples []prompb.Sample) []*pack.TagCountPack {
	ret := make([]*pack.TagCountPack, len(samples))
	for i := range ret {
		p := pack.NewTagCountPack()
		p.Category = "prometheus"
		p.Pcode = int64(pcode)
		p.Oid = whash.HashStr(oname)

		ret[i] = p
	}
	for _, p := range ret {
		p.PutTag("oname", oname)
		for _, l := range labels {
			if l.Name == model.MetricNameLabel {
				continue
			}
			p.PutTag(l.Name, l.Value)
		}
	}

	return ret
}

func getDataName(labels []prompb.Label) (string, error) {
	for _, l := range labels {
		if l.Name == model.MetricNameLabel {
			return l.Value, nil
		}
	}

	return "", errors.New("name label is not exist")
}

func handler(w http.ResponseWriter, r *http.Request) {

	req, err := remote.DecodeWriteRequest(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, ts := range req.Timeseries {
		dataName, err := getDataName(ts.Labels)
		if err != nil {
			panic(err)
		}
		packs := genPackTemplates(ts.Labels, ts.Samples)

		for i, sample := range ts.Samples {
			packs[i].Time = ts.Samples[i].Timestamp
			packs[i].Put(dataName, sample.Value)
		}

		for _, p := range packs {
			err := tcpClient.SendFlush(p, true)
			if err != nil {
				panic(err)
			}
		}
	}

	if err != nil {
		log.Panic(err)
	}

}

func main() {
	flag.StringVar(&license, "license", "", "whatap license")
	flag.Int64Var(&pcode, "pcode", 0, "whatap pcode")
	flag.StringVar(&oname, "oname", "skynet", "agent oname")
	flag.StringVar(&whatapHost, "whatap-host", "13.209.172.35", "whatap host")
	flag.StringVar(&listen, "listen", "0.0.0.0:19090", "listen address")
	flag.Parse()
	servers := make([]string, 0)
	//servers = append(servers, fmt.Sprintf("%s:%d", "13.124.11.223", 6600))
	servers = append(servers, fmt.Sprintf("%s:%d", whatapHost, 6600))
	tcpClient = oneway.GetOneWayTcpClient(
		oneway.WithServers(servers),
		oneway.WithLicense(license),
		oneway.WithPcode(pcode),
		oneway.WithOid(whash.HashStr(oname)),
		oneway.WithUseQueue(),
		oneway.WithLogger(logger.NewDefaultLogger()),
	)
	defer tcpClient.Close()

	http.HandleFunc("/receive", handler)
	log.Fatal("listen", listen, http.ListenAndServe(listen, nil))
}

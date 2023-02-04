package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/whatap/golib/net/oneway"

	"github.com/whatap/golib/logger"
	whash "github.com/whatap/golib/util/hash"

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

func handler(w http.ResponseWriter, r *http.Request) {

	req, err := remote.DecodeWriteRequest(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, ts := range req.Timeseries {
		packs, err := generatePacksFromTimeSeries(ts)
		if err != nil {
			panic(err)
		}
		err = sendPacks(packs)
		if err != nil {
			panic(err)
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

	servers := []string{fmt.Sprintf("%s:%d", whatapHost, 6600)}
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

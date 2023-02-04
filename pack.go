package main

import (
	"errors"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/whatap/golib/lang/pack"
	whash "github.com/whatap/golib/util/hash"
)

func newBasePack() *pack.TagCountPack {
	p := pack.NewTagCountPack()
	p.Category = "prometheus"
	p.Pcode = int64(pcode)
	p.Oid = whash.HashStr(oname)
	p.PutTag("oname", oname)

	return p
}

func newLabeledPack(labels []prompb.Label) *pack.TagCountPack {
	p := newBasePack()
	for _, l := range labels {
		if l.Name == model.MetricNameLabel {
			continue
		}
		p.PutTag(l.Name, l.Value)
	}

	return p
}

func getLabeledPacks(labels []prompb.Label, samples []prompb.Sample) []*pack.TagCountPack {
	ret := make([]*pack.TagCountPack, len(samples))
	for i := range ret {
		ret[i] = newLabeledPack(labels)
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

func generatePacksFromTimeSeries(ts prompb.TimeSeries, metadata prompb.MetricMetadata) ([]*pack.TagCountPack, error) {
	dataName, err := getDataName(ts.Labels)
	if err != nil {
		return nil, err
	}

	packs := getLabeledPacks(ts.Labels, ts.Samples)

	for i, sample := range ts.Samples {
		packs[i].Time = ts.Samples[i].Timestamp
		packs[i].Put(dataName, NewMetric(sample, metadata).Value())
	}

	return packs, nil
}

func sendPacks(packs []*pack.TagCountPack) error {
	for _, p := range packs {
		err := tcpClient.SendFlush(p, true)
		if err != nil {
			return err
		}
	}
	return nil
}

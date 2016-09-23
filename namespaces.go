package main

import (
	"log"
	"strings"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// NamespaceMetrics lists the keys we report from aero's namespace statistics command.
	// See `asinfo -l -v namespace/<namespace>` for the full list.
	NamespaceMetrics = []metric{
		{collGauge, "migrate_rx_partitions_remaining", "remaining rx migrate partitions per namespace per node"},
		{collGauge, "migrate_tx_partitions_remaining", "remaining tx migrate partitions per namespace per node"},
		{collGauge, "memory_free_pct", "% free memory per namespace per node"},
		{collGauge, "device_free_pct", "% free memory per namespace per node"},
		{collGauge, "device_available_pct", "% available pct per namespace per node"},
		{collGauge, "evicted_objects", "evicted objects per namespace per node"},
		{collGauge, "expired_objects", "expired objects per namespace per node"},
		{collGauge, "client_read_success", "reads per namespace per node"},
		{collGauge, "client_write_success", "writes per namespace per node"},
		{collGauge, "client_write_error", "writes error per namespace per node"},
		{collGauge, "client_read_error", "read errors per namespace per node"},
		{collGauge, "client_read_timeout", "read timeout per namespace per node"},
		{collGauge, "client_write_timeout", "write timeout per namespace per node"},
		{collGauge, "objects", "objects per namespace per node"},
	}
)

type nsCollector struct {
	// gauges map[string]*prometheus.GaugeVec
	descs   []prometheus.Collector
	metrics map[string]func(ns string) setter
}

func newNSCollector() *nsCollector {
	var (
		descs   []prometheus.Collector
		metrics = map[string]func(ns string) setter{}
	)
	for _, s := range NamespaceMetrics {
		key := s.aeroName
		promName := strings.Replace(key, "-", "_", -1)
		switch s.typ {
		case collGauge:
			v := prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: systemNamespace,
				Name:      promName,
				Help:      s.desc,
			},
				[]string{"namespace"},
			)
			metrics[key] = func(ns string) setter {
				return v.WithLabelValues(ns)
			}
			descs = append(descs, v)
		case collCounter:
			// todo
		}
	}

	return &nsCollector{
		descs:   descs,
		metrics: metrics,
	}
}

func (c *nsCollector) describe(ch chan<- *prometheus.Desc) {
	for _, d := range c.descs {
		d.Describe(ch)
	}
}

func (c *nsCollector) collect(conn *as.Connection, ch chan<- prometheus.Metric) {
	info, err := as.RequestInfo(conn, "namespaces")
	if err != nil {
		log.Print(err)
		return
	}
	for _, ns := range strings.Split(info["namespaces"], ";") {
		nsinfo, err := as.RequestInfo(conn, "namespace/"+ns)
		if err != nil {
			log.Print(err)
			continue
		}
		ms := map[string]setter{}
		for key, m := range c.metrics {
			ms[key] = m(ns)
		}
		infoCollect(ch, ms, nsinfo["namespace/"+ns])
	}
}

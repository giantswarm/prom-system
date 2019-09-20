package main

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
)

var (
	metricsToAdd    []prometheus.Metric
	metricsToDelete []prometheus.Metric

	replicasetDesiredDesc = prometheus.NewDesc(
		"replicaset_desired",
		"",
		[]string{"name", "image", "replicaset"},
		nil,
	)
	podDesiredDesc = prometheus.NewDesc(
		"pod_desired",
		"",
		[]string{"name", "image", "replicaset"},
		nil,
	)
)

type ServiceCollector struct {
	client v1.API
}

func sampleToStrings(sample model.Metric) []string {
	s := []string{}

	switch sample["__name__"] {
	case "replicaset_desired":
		s = append(s, string(sample["name"]))
		s = append(s, string(sample["image"]))
		s = append(s, string(sample["replicaset"]))
	case "pod_desired":
		s = append(s, string(sample["name"]))
		s = append(s, string(sample["image"]))
		s = append(s, string(sample["replicaset"]))
	}

	return s
}

func (c ServiceCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c ServiceCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	value, _, err := c.client.QueryRange(
		ctx,
		`{__name__=~"replicaset_desired|pod_desired"}`,
		v1.Range{
			Start: time.Now().Add(time.Minute * -1),
			End:   time.Now(),
			Step:  time.Second,
		},
	)
	if err != nil {
		log.Fatalf("Could not query desired state: %v", err)
		return
	}

	m := value.(model.Matrix)

	metrics := []prometheus.Metric{}

	for _, samplestream := range m {
		sample := samplestream
		var metric prometheus.Metric

		switch sample.Metric["__name__"] {
		case "replicaset_desired":
			metric = prometheus.MustNewConstMetric(
				replicasetDesiredDesc, prometheus.GaugeValue, float64(sample.Values[len(sample.Values)-1].Value), sampleToStrings(sample.Metric)...,
			)
		case "pod_desired":
			metric = prometheus.MustNewConstMetric(
				podDesiredDesc, prometheus.GaugeValue, float64(sample.Values[len(sample.Values)-1].Value), sampleToStrings(sample.Metric)...,
			)
		}

		metrics = append(metrics, metric)
	}

	for i, newMetric := range metricsToAdd {
		metricFlushed := false

		for _, existingMetric := range metrics {
			if reflect.DeepEqual(newMetric, existingMetric) {
				metricFlushed = true
				break
			}
		}

		if metricFlushed {
			fmt.Println("New metric to add flushed, removing from queue")
			metricsToAdd = append(metricsToAdd[:i], metricsToAdd[i+1:]...)
		} else {
			fmt.Println("New metric to add not yet flushed, adding to output")
			metrics = append(metrics, newMetric)
		}
	}

	for i, newMetric := range metricsToDelete {
		metricFlushed := true

		for j, existingMetric := range metrics {
			if reflect.DeepEqual(newMetric, existingMetric) {
				metricFlushed = false
				metrics = append(metrics[:j], metrics[j+1:]...)
			}
		}

		if metricFlushed {
			fmt.Println("New metric to delete flushed, removing from queue")
			metricsToDelete = append(metricsToDelete[:i], metricsToDelete[i+1:]...)
		} else {
			fmt.Println("New metric to delete not yet flushed, removed from output")
		}
	}

	for _, metric := range metrics {
		ch <- metric
	}
}

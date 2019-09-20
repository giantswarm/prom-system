package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	podActual = prometheus.NewDesc(
		prometheus.BuildFQName("pod", "", "actual"),
		"Actual state of pods.",
		[]string{"name", "image", "id"},
		nil,
	)
)

type Request struct {
	Alerts []Alert `json:"alerts"`
}

type Alert struct {
	Labels map[string]string `json:"labels"`
}

type PodCollector struct {
}

func (c PodCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c PodCollector) Collect(ch chan<- prometheus.Metric) {
	cmd := exec.Command(
		"docker",
		"ps",
		"--format",
		"{{.Names}} {{.Image}} {{.ID}}",
	)

	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Could not collect Docker metrics:", err)
		return
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		fields := strings.Fields(line)

		if len(fields) == 0 {
			return
		}

		ch <- prometheus.MustNewConstMetric(
			podActual, prometheus.GaugeValue, 1, fields...,
		)
	}
}

func startContainer(name string, image string) error {
	fmt.Println("Starting:", name, image)

	cmd := exec.Command(
		"docker",
		"run",
		"--rm",
		"--name",
		name,
		"--detach",
		image,
	)
	if err := cmd.Run(); err != nil {
		fmt.Println("Could not start container:", err)
		return err
	}

	return nil
}

func stopContainer(name string) error {
	fmt.Println("Stopping:", name)

	cmd := exec.Command(
		"docker",
		"kill",
		name,
	)
	if err := cmd.Run(); err != nil {
		fmt.Println("Could not stop container:", err)
		return err
	}

	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received alert from Alertmanager")

	var req Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println("Could not decode request:", err)
		http.Error(w, err.Error(), 400)
		return
	}

	for _, alert := range req.Alerts {
		var err error

		fmt.Println("Alert name:", alert.Labels["alertname"])

		switch alert.Labels["alertname"] {
		case "PodNotRunning":
			err = startContainer(alert.Labels["name"], alert.Labels["image"])
		case "PodRunningButShouldnt":
			err = stopContainer(alert.Labels["name"])
		}

		if err != nil {
			http.Error(w, err.Error(), 400)
		}
	}
}

func main() {
	fmt.Println("Initialising pod collector")
	collector := PodCollector{}
	prometheus.MustRegister(collector)

	fmt.Println("Starting pod controller")

	http.HandleFunc("/", handler)
	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(":8000", nil))
}

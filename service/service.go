package main

import (
	"log"
	"net"
	"net/http"

	pb "github.com/giantswarm/prom-system/grpc/prom-system"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	client, err := api.NewClient(api.Config{
		Address: "http://127.0.0.1:9090",
	})
	if err != nil {
		log.Fatalf("Could not create Prometheus client: %v", err)
	}

	api := v1.NewAPI(client)

	containerService := containerService{
		client: api,
	}
	replicasetService := replicasetService{
		client: api,
	}

	lis, err := net.Listen("tcp", ":8010")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()

	pb.RegisterContainerServiceServer(server, containerService)
	pb.RegisterReplicasetServiceServer(server, replicasetService)

	go func() {
		log.Println("Starting metrics server")
		serviceCollector := ServiceCollector{
			client: api,
		}
		prometheus.MustRegister(serviceCollector)

		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":8011", nil))
	}()

	log.Println("Starting grpc server")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

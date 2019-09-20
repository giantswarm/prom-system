package main

import (
	"context"
	"fmt"
	"time"

	pb "github.com/giantswarm/prom-system/grpc/prom-system"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
)

type containerService struct {
	client v1.API
}

func valueToContainers(value model.Value) []*pb.Container {
	vector := value.(model.Vector)

	containers := []*pb.Container{}

	for _, sample := range vector {
		container := &pb.Container{}

		for name, value := range sample.Metric {
			t := string(value)

			switch name {
			case "id":
				container.Id = &t
			case "image":
				container.Image = &t
			case "name":
				container.Name = &t
			case "replicaset":
				container.Replicaset = &t
			}
		}

		containers = append(containers, container)
	}

	return containers
}

func (s containerService) CreateContainer(ctx context.Context, req *pb.CreateContainerRequest) (*pb.CreateContainerResponse, error) {
	metric, err := prometheus.NewConstMetric(
		podDesiredDesc,
		prometheus.GaugeValue,
		float64(1),
		[]string{req.Container.GetName(), req.Container.GetImage(), req.Container.GetReplicaset()}...,
	)
	if err != nil {
		return &pb.CreateContainerResponse{}, err
	}

	fmt.Println("Adding metric for:", req.Container.GetName(), req.Container.GetImage())
	metricsToAdd = append(metricsToAdd, metric)

	return &pb.CreateContainerResponse{}, nil
}

func (s containerService) DeleteContainer(ctx context.Context, req *pb.DeleteContainerRequest) (*pb.DeleteContainerResponse, error) {
	metric, err := prometheus.NewConstMetric(
		podDesiredDesc,
		prometheus.GaugeValue,
		float64(1),
		[]string{req.Container.GetName(), req.Container.GetImage(), req.Container.GetReplicaset()}...,
	)
	if err != nil {
		return &pb.DeleteContainerResponse{}, err
	}

	fmt.Println("Deleting metric for:", req.Container.GetName(), req.Container.GetImage())
	metricsToDelete = append(metricsToDelete, metric)

	return &pb.DeleteContainerResponse{}, nil
}

func (s containerService) ListContainers(ctx context.Context, req *pb.ListContainersRequest) (*pb.ListContainersResponse, error) {
	desiredValue, _, err := s.client.Query(ctx, "pod_desired", time.Now())
	if err != nil {
		return &pb.ListContainersResponse{}, err
	}

	actualValue, _, err := s.client.Query(ctx, "pod_actual", time.Now())
	if err != nil {
		return &pb.ListContainersResponse{}, err
	}

	desiredContainers := valueToContainers(desiredValue)
	actualContainers := valueToContainers(actualValue)

	containers := []*pb.Container{}

	for _, desiredContainer := range desiredContainers {
		container := desiredContainer

		for _, actualContainer := range actualContainers {
			if *container.Name == *actualContainer.Name {
				container.Id = actualContainer.Id
				break
			}
		}

		containers = append(containers, container)
	}

	response := &pb.ListContainersResponse{
		Containers: containers,
	}

	return response, nil
}

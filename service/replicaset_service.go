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

type replicasetService struct {
	client v1.API
}

func (s replicasetService) valueToReplicasets(value model.Value) []*pb.Replicaset {
	vector := value.(model.Vector)

	replicasets := []*pb.Replicaset{}

	for _, sample := range vector {
		replicaset := &pb.Replicaset{}

		for name, value := range sample.Metric {
			t := string(value)

			switch name {
			case "image":
				replicaset.Image = &t
			case "name":
				replicaset.Name = &t
			}
		}

		v := int32(sample.Value)
		pods := int32(0)

		replicaset.Replicas = &v
		replicaset.Pods = &pods

		replicasets = append(replicasets, replicaset)
	}

	return replicasets
}

func (s replicasetService) getPodsForReplicaset(ctx context.Context, name string, image string) (int32, error) {
	query := fmt.Sprintf(`sum(pod_desired{replicaset="%v", image="%v"})`, name, image)

	v, _, err := s.client.Query(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}

	vector := v.(model.Vector)
	if len(vector) == 0 {
		return 0, nil
	}

	p := int32(vector[0].Value)

	return p, nil
}

func (s replicasetService) CreateReplicaset(ctx context.Context, req *pb.CreateReplicasetRequest) (*pb.CreateReplicasetResponse, error) {
	metric, err := prometheus.NewConstMetric(
		replicasetDesiredDesc,
		prometheus.GaugeValue,
		float64(req.Replicaset.GetReplicas()),
		[]string{req.Replicaset.GetName(), req.Replicaset.GetImage(), req.Replicaset.GetName()}...,
	)
	if err != nil {
		return &pb.CreateReplicasetResponse{}, err
	}

	fmt.Println("Adding metric for:", req.Replicaset.GetName(), req.Replicaset.GetImage())
	metricsToAdd = append(metricsToAdd, metric)

	return &pb.CreateReplicasetResponse{}, nil
}

func (s replicasetService) DeleteReplicaset(ctx context.Context, req *pb.DeleteReplicasetRequest) (*pb.DeleteReplicasetResponse, error) {
	metric, err := prometheus.NewConstMetric(
		replicasetDesiredDesc,
		prometheus.GaugeValue,
		float64(req.Replicaset.GetReplicas()),
		[]string{req.Replicaset.GetName(), req.Replicaset.GetImage(), req.Replicaset.GetName()}...,
	)
	if err != nil {
		return &pb.DeleteReplicasetResponse{}, err
	}

	fmt.Println("Deleting metric for:", req.Replicaset.GetName(), req.Replicaset.GetImage())
	metricsToDelete = append(metricsToDelete, metric)

	return &pb.DeleteReplicasetResponse{}, nil
}

func (s replicasetService) ListReplicasets(ctx context.Context, req *pb.ListReplicasetsRequest) (*pb.ListReplicasetsResponse, error) {
	desiredValue, _, err := s.client.Query(ctx, "replicaset_desired", time.Now())
	if err != nil {
		return &pb.ListReplicasetsResponse{}, err
	}

	replicasets := s.valueToReplicasets(desiredValue)

	for _, replicaset := range replicasets {
		p, err := s.getPodsForReplicaset(ctx, *replicaset.Name, *replicaset.Image)
		if err != nil {
			return &pb.ListReplicasetsResponse{}, err
		}
		replicaset.Pods = &p
	}

	response := &pb.ListReplicasetsResponse{
		Replicasets: replicasets,
	}

	return response, nil
}

func (s replicasetService) ScaleReplicaset(ctx context.Context, req *pb.ScaleReplicasetRequest) (*pb.ScaleReplicasetResponse, error) {
	oldMetric, err := prometheus.NewConstMetric(
		replicasetDesiredDesc,
		prometheus.GaugeValue,
		float64(req.Replicaset.GetReplicas()),
		[]string{req.Replicaset.GetName(), req.Replicaset.GetImage(), req.Replicaset.GetName()}...,
	)
	if err != nil {
		return &pb.ScaleReplicasetResponse{}, err
	}

	newMetric, err := prometheus.NewConstMetric(
		replicasetDesiredDesc,
		prometheus.GaugeValue,
		float64(req.GetNewReplicas()),
		[]string{req.Replicaset.GetName(), req.Replicaset.GetImage(), req.Replicaset.GetName()}...,
	)
	if err != nil {
		return &pb.ScaleReplicasetResponse{}, err
	}

	fmt.Println("Update metric for:", req.Replicaset.GetName(), req.Replicaset.GetImage())
	metricsToAdd = append(metricsToAdd, newMetric)
	metricsToDelete = append(metricsToDelete, oldMetric)

	return &pb.ScaleReplicasetResponse{}, nil
}

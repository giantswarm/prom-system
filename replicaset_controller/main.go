package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	pb "github.com/giantswarm/prom-system/grpc/prom-system"
	"google.golang.org/grpc"
)

const (
	server = "localhost:8010"
)

type Request struct {
	Alerts []Alert `json:"alerts"`
}

type Alert struct {
	Labels map[string]string `json:"labels"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")
)

func createPodDesired(name string, image string) error {
	fmt.Println("Adding container desired:", name, image)

	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Printf("Could not dial server: %v", err)
		return err
	}

	client := pb.NewContainerServiceClient(conn)

	ctx := context.Background()

	id := fmt.Sprintf("%v-%v", name, RandStringRunes(5))

	req := &pb.CreateContainerRequest{
		Container: &pb.Container{
			Name:       &id,
			Image:      &image,
			Replicaset: &name,
		},
	}
	if _, err := client.CreateContainer(ctx, req); err != nil {
		log.Printf("Could not create container: %v", err)
		return err
	}

	return nil
}

func removePodDesired(name string, image string) error {
	fmt.Println("Removing pod desired:", name, image)

	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Printf("Could not dial server: %v", err)
		return err
	}

	client := pb.NewContainerServiceClient(conn)

	ctx := context.Background()

	resp, err := client.ListContainers(ctx, &pb.ListContainersRequest{})
	if err != nil {
		log.Printf("Could not list containers: %v", err)
		return err
	}

	for _, container := range resp.Containers {
		fmt.Println(name, image)
		fmt.Println(container)

		if container.GetReplicaset() == name && container.GetImage() == image {
			req := &pb.DeleteContainerRequest{
				Container: container,
			}
			if _, err := client.DeleteContainer(ctx, req); err != nil {
				log.Printf("Could not delete container: %v", err)
				return err
			}

			break
		}
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
		case "ReplicasetUnderSatisfied":
			err = createPodDesired(alert.Labels["name"], alert.Labels["image"])
		case "ReplicasetOverSatisfied":
			err = removePodDesired(alert.Labels["name"], alert.Labels["image"])
		}

		if err != nil {
			http.Error(w, err.Error(), 400)
		}
	}
}

func main() {
	fmt.Println("Starting replicaset controller")

	http.HandleFunc("/", handler)

	log.Fatal(http.ListenAndServe(":8001", nil))
}

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

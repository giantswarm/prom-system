package cmd

import (
	"context"
	"fmt"
	"log"
	"strconv"

	pb "github.com/giantswarm/prom-system/grpc/prom-system"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	scaleReplicasetCmd = &cobra.Command{
		Use:   "scale-replicaset",
		Short: "Scale a replicaset",
		Run:   executeScaleRelicasetCmd,
	}
)

func init() {
	rootCmd.AddCommand(scaleReplicasetCmd)
}

func executeScaleRelicasetCmd(cmd *cobra.Command, args []string) {
	if len(args) != 4 {
		log.Fatalf("Please specify name, image, number of replicas, and new number of replicas")
	}
	name := args[0]
	image := args[1]
	r, err := strconv.Atoi(args[2])
	if err != nil {
		log.Fatalf("Invalid number of replicas: %v", err)
	}
	replicas := int32(r)
	nr, err := strconv.Atoi(args[3])
	if err != nil {
		log.Fatalf("Invalid number of new replicas: %v", err)
	}
	newReplicas := int32(nr)

	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not dial server: %v", err)
	}

	client := pb.NewReplicasetServiceClient(conn)

	ctx := context.Background()

	req := &pb.ScaleReplicasetRequest{
		Replicaset: &pb.Replicaset{
			Name:     &name,
			Image:    &image,
			Replicas: &replicas,
		},
		NewReplicas: &newReplicas,
	}
	if _, err := client.ScaleReplicaset(ctx, req); err != nil {
		log.Fatalf("Could not scale replicaset: %v", err)
	}

	fmt.Println("Replicaset scaled!")
}

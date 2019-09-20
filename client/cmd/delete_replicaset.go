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
	deleteReplicasetCmd = &cobra.Command{
		Use:   "delete-replicaset",
		Short: "Delete a replicaset",
		Run:   executeDeleteRelicasetCmd,
	}
)

func init() {
	rootCmd.AddCommand(deleteReplicasetCmd)
}

func executeDeleteRelicasetCmd(cmd *cobra.Command, args []string) {
	if len(args) != 3 {
		log.Fatalf("Please specify name, image, and number of replicas")
	}
	name := args[0]
	image := args[1]
	r, err := strconv.Atoi(args[2])
	if err != nil {
		log.Fatalf("Invalid number of replicas: %v", err)
	}
	replicas := int32(r)

	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not dial server: %v", err)
	}

	client := pb.NewReplicasetServiceClient(conn)

	ctx := context.Background()

	req := &pb.DeleteReplicasetRequest{
		Replicaset: &pb.Replicaset{
			Name:     &name,
			Image:    &image,
			Replicas: &replicas,
		},
	}
	if _, err := client.DeleteReplicaset(ctx, req); err != nil {
		log.Fatalf("Could not delete replicaset: %v", err)
	}

	fmt.Println("Replicaset deleted!")
}

package cmd

import (
	"context"
	"fmt"
	"log"

	pb "github.com/giantswarm/prom-system/grpc/prom-system"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	deleteContainerCmd = &cobra.Command{
		Use:   "delete-container",
		Short: "Delete a container",
		Run:   executeDeleteContainerCmd,
	}
)

func init() {
	rootCmd.AddCommand(deleteContainerCmd)
}

func executeDeleteContainerCmd(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		log.Fatalf("Please specify name and image")
	}
	name := args[0]
	image := args[1]
	replicaset := ""
	if len(args) == 3 {
		replicaset = args[2]
	}

	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not dial server: %v", err)
	}

	client := pb.NewContainerServiceClient(conn)

	ctx := context.Background()

	req := &pb.DeleteContainerRequest{
		Container: &pb.Container{
			Name:  &name,
			Image: &image,
		},
	}
	if replicaset != "" {
		req.Container.Replicaset = &replicaset
	}

	if _, err := client.DeleteContainer(ctx, req); err != nil {
		log.Fatalf("Could not delete container: %v", err)
	}

	fmt.Println("Container deleted!")
}

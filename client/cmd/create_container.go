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
	createContainerCmd = &cobra.Command{
		Use:   "create-container",
		Short: "Create a container",
		Run:   executeCreateContainerCmd,
	}
)

func init() {
	rootCmd.AddCommand(createContainerCmd)
}

func executeCreateContainerCmd(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		log.Fatalf("Please specify name and image")
	}
	name := args[0]
	image := args[1]

	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not dial server: %v", err)
	}

	client := pb.NewContainerServiceClient(conn)

	ctx := context.Background()

	req := &pb.CreateContainerRequest{
		Container: &pb.Container{
			Name:  &name,
			Image: &image,
		},
	}
	if _, err := client.CreateContainer(ctx, req); err != nil {
		log.Fatalf("Could not create container: %v", err)
	}

	fmt.Println("Container created!")
}

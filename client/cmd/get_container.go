package cmd

import (
	"context"
	"fmt"
	"log"

	pb "github.com/giantswarm/prom-system/grpc/prom-system"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	getContainerCmd = &cobra.Command{
		Use:   "get-container",
		Short: "Get all containers",
		Run:   executeContainerCmd,
	}
)

func init() {
	rootCmd.AddCommand(getContainerCmd)
}

func executeContainerCmd(cmd *cobra.Command, args []string) {
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not dial server: %v", err)
	}

	client := pb.NewContainerServiceClient(conn)

	ctx := context.Background()
	resp, err := client.ListContainers(ctx, &pb.ListContainersRequest{})
	if err != nil {
		log.Fatalf("Could not list containers: %v", err)
	}

	output := []string{"NAME | IMAGE | REPLICASET | STATUS"}
	for _, container := range resp.Containers {
		ready := "PENDING"
		if container.GetId() != "" {
			ready = "RUNNING"
		}

		s := fmt.Sprintf("%v | %v | %v | %v", container.GetName(), container.GetImage(), container.GetReplicaset(), ready)
		output = append(output, s)
	}

	result := columnize.SimpleFormat(output)

	fmt.Println(result)
}

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
	getReplicasetCmd = &cobra.Command{
		Use:   "get-replicaset",
		Short: "Get all replicasets",
		Run:   executeReplicasetCmd,
	}
)

func init() {
	rootCmd.AddCommand(getReplicasetCmd)
}

func executeReplicasetCmd(cmd *cobra.Command, args []string) {
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not dial server: %v", err)
	}

	client := pb.NewReplicasetServiceClient(conn)

	ctx := context.Background()
	resp, err := client.ListReplicasets(ctx, &pb.ListReplicasetsRequest{})
	if err != nil {
		log.Fatalf("Could not list replicasets: %v", err)
	}

	output := []string{"NAME | IMAGE | REPLICAS | PODS"}
	for _, replicaset := range resp.Replicasets {
		s := fmt.Sprintf("%v | %v | %v | %v", replicaset.GetName(), replicaset.GetImage(), replicaset.GetReplicas(), replicaset.GetPods())
		output = append(output, s)
	}

	result := columnize.SimpleFormat(output)

	fmt.Println(result)
}

package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

const (
	server = "localhost:8010"
)

var (
	rootCmd = &cobra.Command{
		Use: "prom-system",
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("%v\n", err)
	}
}

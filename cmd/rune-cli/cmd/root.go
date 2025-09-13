package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	apiv1 "github.com/thelamedev/rune/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	client apiv1.RuneServiceClient

	rootCmd = &cobra.Command{
		Use:   "rune-cli",
		Short: "A CLI to interacct with a Rune server",
		Long:  `rune-cli is a command-line interface to the Rune secrets management and service discovery system.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			conn, err := grpc.Dial("localhost:8000", grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to connect to Rune server: %v\n", err)
				os.Exit(1)
			}

			client = apiv1.NewRuneServiceClient(conn)
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

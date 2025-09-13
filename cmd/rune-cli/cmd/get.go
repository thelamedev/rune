package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	apiv1 "github.com/thelamedev/rune/api/v1"
)

var getCmd = &cobra.Command{
	Use:   "get [path]",
	Short: "Get a secret at a given path",
	Long:  `Retrieves a secret value at a specified path in the Rune vault.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]

		resp, err := client.Get(cmd.Context(), &apiv1.GetRequest{
			Path: path,
		})
		if err != nil {
			fmt.Printf("Failed to get secret: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Secret retrieved at %q: %s\n", path, resp.Value)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	apiv1 "github.com/thelamedev/rune/api/v1"
)

var putCmd = &cobra.Command{
	Use:   "put [path] [value]",
	Short: "Put a secret at a given path",
	Long:  `Stores a secret value at a specified path in the Rune vault.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		value := args[1]

		_, err := client.Put(cmd.Context(), &apiv1.PutRequest{
			Path:  path,
			Value: []byte(value),
		})
		if err != nil {
			fmt.Printf("Failed to put secret: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Secret stored at %q\n", path)
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
}

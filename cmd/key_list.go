package cmd

import (
	"fmt"

	"github.com/sakkshm/bastion/internal/auth"
	"github.com/spf13/cobra"
)

var keyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API Keys",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing keys...")
		err := auth.ListAPIKeys()
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	keyCmd.AddCommand(keyListCmd)
}

package cmd

import (
	"fmt"

	"github.com/sakkshm/bastion/internal/auth"
	"github.com/spf13/cobra"
)

var public_id string

var keyRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Remove API keys",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Revoking key...")
		err := auth.RevokeAPIKey(public_id)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	keyRevokeCmd.Flags().StringVar(&public_id, "public_id", "", "Public ID of API key")
	keyCmd.AddCommand(keyRevokeCmd)
}

package cmd

import (
	"fmt"

	"github.com/sakkshm/bastion/internal/auth"
	"github.com/spf13/cobra"
)

var name string
var scope string

var keyCreateCmd = &cobra.Command{
	Use: "create",
	Short: "Create API keys",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Creating key: %s with scope: %s", name, scope)
		err := auth.CreateAPIKey(name, scope)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	keyCreateCmd.Flags().StringVar(&name, "name", "default", "Key name")
	keyCreateCmd.Flags().StringVar(&scope, "scope", "viewer", "Scope of Key")
	keyCmd.AddCommand(keyCreateCmd)
}
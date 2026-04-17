package cmd

import "github.com/spf13/cobra"

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage API keys",
}

func init() {
	RootCmd.AddCommand(keyCmd)
}

package cmd

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "bastion",
	Short: "Bastion CLI - Sandboxed Runtime Env.",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func Execute() {
	cobra.CheckErr(RootCmd.Execute())
}

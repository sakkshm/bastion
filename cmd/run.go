package cmd

import (
	"github.com/sakkshm/bastion/internal/bastion"
	"github.com/spf13/cobra"
)

var configPath string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the Bastion Runtime",
	Run: func(cmd *cobra.Command, args []string) {
		bastion.RunBastion(configPath)
	},
}

func init() {
	runCmd.Flags().StringVar(&configPath, "config", "./config/config", "Path to Config.toml")
	RootCmd.AddCommand(runCmd)
}

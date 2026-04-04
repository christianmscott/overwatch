package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "overwatch",
	Short: "Monitoring tool for sites, services, jobs, and more",
	Long:  "Overwatch is an open source monitoring tool that runs health checks and alerts on failures.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: overwatch.yaml)")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(checksCmd)
	rootCmd.AddCommand(alertsCmd)
	rootCmd.AddCommand(versionCmd)
}

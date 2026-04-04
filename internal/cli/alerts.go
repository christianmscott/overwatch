package cli

import (
	"fmt"

	"github.com/christianmscott/overwatch/internal/alerts"
	"github.com/christianmscott/overwatch/internal/config"
	"github.com/spf13/cobra"
)

var alertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Manage and test alerts",
}

var alertsTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Send a test alert through all configured senders",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := cfgFile
		if path == "" {
			path = config.DefaultPath
		}

		cfg, err := config.Load(path)
		if err != nil {
			return err
		}

		senders := alerts.BuildSenders(cfg.Alerts)
		if len(senders) == 0 {
			return fmt.Errorf("no alert senders configured; add webhooks or smtp to your config")
		}

		router := alerts.NewRouter(senders)
		router.SendTest()

		fmt.Printf("test alert sent to %d sender(s)\n", len(senders))
		return nil
	},
}

func init() {
	alertsCmd.AddCommand(alertsTestCmd)
}

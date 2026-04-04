package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/christianmscott/overwatch/internal/checks"
	"github.com/christianmscott/overwatch/internal/config"
	"github.com/spf13/cobra"
)

var checksCmd = &cobra.Command{
	Use:   "checks",
	Short: "Manage and test checks",
}

var checksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured checks",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := cfgFile
		if path == "" {
			path = config.DefaultPath
		}

		cfg, err := config.Load(path)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tTARGET\tINTERVAL\tTIMEOUT")
		for _, c := range cfg.Checks {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", c.Name, c.Type, c.Target, c.Interval.Duration, c.Timeout.Duration)
		}
		return w.Flush()
	},
}

var checksTestCmd = &cobra.Command{
	Use:   "test [name]",
	Short: "Run a check by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := cfgFile
		if path == "" {
			path = config.DefaultPath
		}

		cfg, err := config.Load(path)
		if err != nil {
			return err
		}

		name := args[0]
		for _, c := range cfg.Checks {
			if c.Name == name {
				result := checks.Run(cmd.Context(), c)
				fmt.Printf("check:    %s\n", result.CheckName)
				fmt.Printf("status:   %s\n", result.Status)
				fmt.Printf("duration: %s\n", result.Duration)
				if result.Error != "" {
					fmt.Printf("error:    %s\n", result.Error)
				}
				if result.Status == "down" {
					os.Exit(1)
				}
				return nil
			}
		}

		return fmt.Errorf("check %q not found in config", name)
	},
}

func init() {
	checksCmd.AddCommand(checksListCmd)
	checksCmd.AddCommand(checksTestCmd)
}

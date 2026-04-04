package cli

import (
	"fmt"
	"os"

	"github.com/christianmscott/overwatch/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a starter configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := cfgFile
		if path == "" {
			path = config.DefaultPath
		}

		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists; remove it first or choose a different path with --config", path)
		}

		if err := os.WriteFile(path, []byte(config.StarterConfig), 0644); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}

		fmt.Printf("wrote starter config to %s\n", path)
		return nil
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := cfgFile
		if path == "" {
			path = config.DefaultPath
		}

		_, err := config.Load(path)
		if err != nil {
			return err
		}

		fmt.Printf("%s is valid\n", path)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configValidateCmd)
}

package cli

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Retrieve the server's join token (requires authentication)",
	Long:  "Fetch the join token from a running Overwatch server. Requires a configured and authenticated client.",
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := serverAddr()
		if err != nil {
			return err
		}
		resp, err := apiDo("GET", addr+"/api/token", nil)
		if err != nil {
			return fmt.Errorf("cannot reach server at %s: %w\nIs 'overwatch serve' running?", addr, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return apiReadError(resp)
		}

		var out struct {
			JoinToken string `json:"join_token"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return err
		}
		fmt.Println(out.JoinToken)
		return nil
	},
}

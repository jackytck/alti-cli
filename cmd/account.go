package cmd

import (
	"os"

	"github.com/jackytck/alti-cli/config"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// accountCmd represents the account command
var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "List all the available accounts",
	Long:  "List all the previously logined accoutns across different servers.",
	Run: func(cmd *cobra.Command, args []string) {
		config := config.Load()
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Endpoint", "Username", "Status"})
		for _, v := range config.Scopes {
			for _, p := range v.Profiles {
				r := []string{p.ID, v.Endpoint, p.Name, ""}
				if config.Active == p.ID {
					r[3] = "Active"
				}
				table.Append(r)
			}
		}
		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(accountCmd)
}

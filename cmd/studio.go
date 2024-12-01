/*
Copyright © 2024 Ebenezer Arthur arthurebenezer@aol.com
*/

package cmd

import (
	"github.com/charmbracelet/glamour"
	"github.com/ebarthur/jotl/cmd/flags"
	"github.com/spf13/cobra"
)

const lngMessage = (`The studio command launches a web-based dashboard for Jotl.

It provides a modern, user-friendly interface to:
- View and filter logs across different environments
- Search and analyze logs by timestamp, status codes, and messages
- Monitor real-time log updates through the dashboard
- Export and share log data

The dashboard automatically starts on port 8080 and will increment
until it finds an available port if 8080 is in use.

Note: The studio dashboard requires the project to be initialized with 'jotl init'
and have a valid database connection configured in the jotl directory.`)

var studioCommand = &cobra.Command{
	Use:   "studio",
	Short: "Start a local web server to view logs",
	Long: func() string {
		out, _ := glamour.Render(lngMessage, "dark")
		return out
	}(),

	Run: func(cmd *cobra.Command, args []string) {
		//
	},
}

func init() {
	var port flags.Port

	studioCommand.Flags().VarP(&port, "port", "p", "Port to run the studio dashboard (default: 8080)")

	// Set default port
	port.Set("8080")

	rootCmd.AddCommand(studioCommand)
}

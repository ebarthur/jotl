package cmd

import (
	"github.com/charmbracelet/glamour"
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
	Short: "Start logging console output to database with optional real-time display",
	Long: func() string {
		out, _ := glamour.Render(lngMessage, "dark")
		return out
	}(),

	Run: func(cmd *cobra.Command, args []string) {
		//
	},
}

func init() {
	rootCmd.AddCommand(studioCommand)
	studioCommand.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the dashboard server")
}

var port int

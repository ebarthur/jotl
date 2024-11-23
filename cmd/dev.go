package cmd

import (
	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
)

const longMsg = (`The dev command starts the core logging functionality of Jotl.

It captures console output and asynchronously logs all output into the configured database.
When running, it will:
- Verify database connection and apply any pending migrations
- Connect to the database and log all console output
- Store structured data including timestamp, environment, message, and status codes

For npm/node projects, add to your package.json scripts:
"dev": "jotl dev --watch & npm run dev"
"start": "jotl dev & npm start"
`)

var devCommand = &cobra.Command{
	Use:   "dev",
	Short: "Start logging console output to database with optional real-time display",
	Long: func() string {
		out, _ := glamour.Render(longMsg, "dark")
		return out
	}(),

	Run: func(cmd *cobra.Command, args []string) {
		//
	},
}

func init() {
	rootCmd.AddCommand(devCommand)

}

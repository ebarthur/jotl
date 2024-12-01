/*
Copyright © 2024 Ebenezer Arthur arthurebenezer@aol.com
*/

package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/glamour"
	"github.com/ebarthur/jotl/cmd/config"
	"github.com/ebarthur/jotl/cmd/db"
	"github.com/ebarthur/jotl/cmd/logs"
	t "github.com/ebarthur/jotl/cmd/types"
	"github.com/ebarthur/jotl/cmd/ui/tui"
	"github.com/ebarthur/jotl/cmd/utils"
	"github.com/spf13/cobra"
)

const longMsg = `The dev command starts the core logging functionality of Jotl.

It captures console output and asynchronously logs all output into the configured database.
When running, it will:
- Verify database connection and apply any pending migrations
- Connect to the database and log all console output
- Store structured data including timestamp, environment, message, and status codes

For npm/node projects, add to your package.json scripts:
"dev": "jotl dev --watch & npm run dev"
"start": "jotl dev & npm start"
`

func init() {
	rootCmd.AddCommand(devCommand)
	devCommand.Flags().BoolP("watch", "w", false, "Open TUI dashboard for real-time logging")
}

var devCommand = &cobra.Command{
	Use:   "dev",
	Short: "Start logging console output to database with optional real-time display",
	Long: func() string {
		out, _ := glamour.Render(longMsg, "dark")
		return out
	}(),

	Run: func(cmd *cobra.Command, args []string) {
		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Println("Failed to get the current directory:", err)
			return
		}

		flagWatch, _ := cmd.Flags().GetBool("watch")

		// Prevent multiple sessions
		lockFile, err := utils.AcquireLock()
		if err != nil {
			fmt.Printf("%s\n", logoStyle.Render(logo))
			fmt.Println(endingMsgStyle.Render("\nAnother instance of `jotl dev` currently running :/"))
			// See docs
			return
		}
		defer func() {
			if err := utils.ReleaseLock(lockFile); err != nil {
				log.Printf("Failed to release lock: %v\n", err)
			}
		}()

		if !utils.IsJotlInitialized(currentDir) {
			fmt.Printf("%s\n", logoStyle.Render(logo))
			fmt.Println(endingMsgStyle.Render("\nYou need to initialize a Jotl project first. Run `jotl init` on your project root directory!"))

			return
		}

		userConfig, err := config.LoadConfig(currentDir)
		if err != nil {
			fmt.Printf("%s\n", logoStyle.Render(logo))
			fmt.Println(endingMsgStyle.Render("\nOops, there is something wrong with your jotl.config.yaml. Verify config and try again."))
			// put link to docs
			return
		}

		logConfig := &t.LogEntry{
			Level:       string(userConfig.Logging.Level),
			ServiceName: userConfig.Project.Name,
			Environment: utils.GetDevEnvironment(),
		}

		dbPath := utils.GetConfigPaths(currentDir).DBDir + "/jotl.db"

		// Initialize the database connection
		logsDB, err := db.NewLogsDB(dbPath)
		if err != nil {
			fmt.Println("Failed to connect to the database. Verify your database configuration and try again.")
			return
		}

		// Initialize the LogStreamer
		logStreamer := logs.NewLogStreamer(logsDB, logConfig)

		if flagWatch {
			tui.ShowDashboard(logStreamer, userConfig)
		} else {
			fmt.Printf("%s\n", logoStyle.Render(logo))

			fmt.Println(endingMsgStyle.Render("Silently logging to your database. Happy hacking!\n"))
			fmt.Println(tipMsgStyle.Render("Tip: Run `jotl dev --watch` to open an interactive TUI dashboard"))

			fmt.Println(endingMsgStyle.Render("Press Ctrl+C to stop the Jotl log stream."))
		}

		// Redirect stdout and stderr before starting streamer
		// This way we are sure to catch any log before starting logger(streamer)
		err = logStreamer.RedirectStd(userConfig)
		if err != nil {
			fmt.Println(err)
		}

		logStreamer.Start()
		defer logStreamer.Stop()

		// Wait for termination signal
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

		// Terminate
		<-signalChannel
	},
}

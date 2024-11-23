/*
Copyright © 2024 Ebenezer Arthur arthurebenezer@aol.com
*/
package cmd

import (
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
)

const lngMsg = (`Jotl is a versatile CLI tool designed to streamline log management for developers by logging console outputs into an SQLite database. 

Jotl comes packed with powerful features:
- A **studio**, which provides an interactive web-based UI for log management.
- A **real-time terminal dashboard (TUI)** for tracking errors and logs directly within your terminal.

Jotl enables developers to initialize projects, log application output in a structured database, and review logs via either the command line or a web UI. Whether you're debugging locally or managing multiple environments, Jotl offers a flexible and reliable solution.

Give us a star on the [repo](https://github.com/ebarthur/jotl)`)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "jotl",
	Short: "Jotl is a CLI tool for logging console outputs into an SQLite database. It features a studio and a real-time (TUI) dashboard that make it easyy to track and manage erros in real-time.",
	Long: func() string {
		out, _ := glamour.Render(lngMsg, "dark")
		return out
	}(),
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

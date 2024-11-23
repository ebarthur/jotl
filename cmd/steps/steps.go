// Package steps provides utility for creating
// each step of the CLI
package steps

import "github.com/ebarthur/jotl/cmd/flags"

// A StepSchema contains the data that is used
// for an individual step of the CLI
type StepSchema struct {
	StepName string // The name of a given step
	Options  []Item // The slice of each option for a given step
	Headers  string // The title displayed at the top of a given step
	Field    string
}

// Steps contains a slice of steps
type Steps struct {
	Steps map[string]StepSchema
}

// An Item contains the data for each option
// in a StepSchema.Options
type Item struct {
	Flag, Title, Desc string
}

// InitSteps initializes and returns the *Steps to be used in the CLI program
func InitSteps(databaseType flags.Database, logLevel flags.LogLevel) *Steps {
	steps := &Steps{
		map[string]StepSchema{
			"driver": {
				StepName: "Database Driver",
				Options: []Item{
					{
						Title: "Sqlite",
						Desc:  "Store logs in a lightweight, file-based SQLite database",
					},
					{
						Title: "Postgres",
						Desc:  "Store logs in a robust, production-ready PostgreSQL database",
					},
				},
				Headers: "What database driver do you want to use in your Jotl project?",
				Field:   databaseType.String(),
			},
			"log_level": {
				StepName: "Log Level",
				Headers:  "Choose log level.",
				Options: []Item{
					{
						Title: "Info",
						Desc:  "Standard information logging",
					},
					{
						Title: "Debug",
						Desc:  "Detailed logging for debugging purposes",
					},
					{
						Title: "Warn",
						Desc:  "Log warning messages and higher severity issues",
					},
					{
						Title: "Error",
						Desc:  "Only log errors and critical issues",
					},
				},
				Field: logLevel.String(),
			},
			"git": {
				StepName: "Git Repository",
				Headers:  "Initialize a Git Repository for your Jotl project.",
				Options: []Item{
					{
						Title: "Yes",
						Desc:  "Initialize a new git repository stage all changes",
					},
					{
						Title: "Skip",
						Desc:  "Proceed without initializing a git repository",
					},
				},
			},
		},
	}

	return steps
}

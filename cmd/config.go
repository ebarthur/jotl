/*
Copyright © 2024 Ebenezer Arthur arthurebenezer@aol.com
*/

package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(configCommand)
}

var configCommand = &cobra.Command{
	Use:   "config",
	Long:  "",
	Short: "",

	// Run: func(cmd *cobra.Command, args []string) {}
}

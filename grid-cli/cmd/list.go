// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available resources on Threefold grid",
}

func init() {
	rootCmd.AddCommand(listCmd)
}

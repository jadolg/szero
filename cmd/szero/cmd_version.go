package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s, Commit: %s, Date: %s, BuiltBy: %s, GoVersion: %s\n", Version, Commit, Date, BuiltBy, GoVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

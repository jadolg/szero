package main

import (
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Version: %s, Commit: %s, Date: %s, BuiltBy: %s GoVersion: %s", Version, Commit, Date, BuiltBy, GoVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

package main

import (
	"fmt"
	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/spf13/cobra"
)

var manpageCmd = &cobra.Command{
	Use:    "manpage",
	Short:  "Generate manpage for szero",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		manPage, err := mcobra.NewManPage(1, rootCmd)
		if err != nil {
			panic(err)
		}

		fmt.Println(manPage.Build(roff.NewDocument()))
	},
}

func init() {
	rootCmd.AddCommand(manpageCmd)
}

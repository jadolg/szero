package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:     "up",
	Short:   "Upscale all deployments in a namespace to their original size",
	Aliases: []string{"upscale"},
	Run: func(cmd *cobra.Command, args []string) {
		clientset, err := getClientset(kubeconfig)
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()
		deployments, err := getDeployments(ctx, clientset, namespace)
		if err != nil {
			log.Fatal(err)
		}

		err = upscaleDeployments(ctx, clientset, deployments)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
}

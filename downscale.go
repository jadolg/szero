package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:     "down",
	Short:   "Downscale all deployments in a namespace",
	Aliases: []string{"downscale"},
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

		log.Infof("Found %d deployments", len(deployments.Items))
		downscaled, err := downscaleDeployments(ctx, clientset, deployments)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Downscaled %d deployments", downscaled)
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}

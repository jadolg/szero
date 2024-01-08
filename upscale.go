package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:     "up",
	Short:   "Upscale all deployments in the desired namespaces to their original size",
	Example: "szero up -n default -n klum",
	Aliases: []string{"upscale"},
	Run: func(cmd *cobra.Command, args []string) {
		clientset, err := getClientset(kubeconfig)
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()
		for _, namespace := range namespaces {
			deployments, err := getDeployments(ctx, clientset, namespace)
			if err != nil {
				log.Fatal(err)
			}

			log.Infof("Found %d deployments in namespace %s", len(deployments.Items), namespace)
			upscaled, err := upscaleDeployments(ctx, clientset, deployments)
			if err != nil {
				log.Fatal(err)
			}
			log.Infof("Upscaled %d deployments", upscaled)
		}
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
}

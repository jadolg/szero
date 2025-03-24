package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:     "down",
	Short:   "Downscale all deployments/statefulsets in the desired namespaces",
	Example: "szero down -n default -n klum",
	Aliases: []string{"downscale"},
	Run: func(cmd *cobra.Command, args []string) {
		clientset, err := getClientset(kubeconfig, kubecontext)
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
			downscaledDeployments, err := downscaleDeployments(ctx, clientset, deployments)
			if err != nil {
				log.Fatal(err)
			}
			log.Infof("Downscaled %d deployments", downscaledDeployments)

			statefulsets, err := getStatefulSets(ctx, clientset, namespace)
			if err != nil {
				log.Fatal(err)
			}

			log.Infof("Found %d statefulsets in namespace %s", len(statefulsets.Items), namespace)
			downscaledStatefulsets, err := downscaleStatefulSets(ctx, clientset, statefulsets)
			if err != nil {
				log.Fatal(err)
			}
			log.Infof("Downscaled %d statefulsets", downscaledStatefulsets)

		}
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}

package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:     "restart",
	Short:   "Restart all deployments/statefulsets in the desired namespaces",
	Example: "szero restart -n default -n klum",
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
			restartedDeployments, err := restartDeployments(ctx, clientset, deployments)
			if err != nil {
				log.Fatal(err)
			}
			log.Infof("Restarted %d deployments", restartedDeployments)

			statefulsets, err := getStatefulSets(ctx, clientset, namespace)
			if err != nil {
				log.Fatal(err)
			}

			log.Infof("Found %d statefulsets in namespace %s", len(statefulsets.Items), namespace)
			restartedStatefulsets, err := restartStatefulSets(ctx, clientset, statefulsets)
			if err != nil {
				log.Fatal(err)
			}
			log.Infof("Restarted %d statefulsets", restartedStatefulsets)
		}
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}

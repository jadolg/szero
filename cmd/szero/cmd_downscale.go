package main

import (
	"context"
	"github.com/jadolg/szero/pkg"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:     "down",
	Short:   "Downscale all deployments/statefulsets/daemonsets in the desired namespaces",
	Example: "szero down -n default -n klum",
	Aliases: []string{"downscale"},
	Run: func(cmd *cobra.Command, args []string) {
		clientset, err := pkg.GetClientset(kubeconfig, kubecontext)
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()
		for _, namespace := range namespaces {
			if skipDeployments {
				log.Infof("Skipping deployments in namespace %s", namespace)
			} else {
				deployments, err := pkg.GetDeployments(ctx, clientset, namespace)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Found %d deployments in namespace %s", len(deployments.Items), namespace)
				downscaledDeployments, err := pkg.DownscaleDeployments(ctx, clientset, deployments)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Downscaled %d deployments", downscaledDeployments)
			}

			if skipStatefulsets {
				log.Infof("Skipping statefulsets in namespace %s", namespace)
			} else {
				statefulsets, err := pkg.GetStatefulSets(ctx, clientset, namespace)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Found %d statefulsets in namespace %s", len(statefulsets.Items), namespace)
				downscaledStatefulsets, err := pkg.DownscaleStatefulSets(ctx, clientset, statefulsets)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Downscaled %d statefulsets", downscaledStatefulsets)
			}

			if skipDaemonsets {
				log.Infof("Skipping daemonsets in namespace %s", namespace)
			} else {
				daemonsets, err := pkg.GetDaemonsets(ctx, clientset, namespace)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Found %d daemonsets in namespace %s", len(daemonsets.Items), namespace)
				downscaleDaemonsets, err := pkg.DownscaleDaemonsets(ctx, clientset, daemonsets)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Downscaled %d daemonsets", downscaleDaemonsets)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}

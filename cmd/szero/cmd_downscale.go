package main

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/jadolg/szero/pkg"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:     "down",
	Short:   "Downscale all deployments/statefulsets/daemonsets in the desired namespaces",
	Example: "szero down -n default -n klum",
	Aliases: []string{"downscale"},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		if dryRun {
			log.Warn("Running in dry-run mode, no changes will be made")
		}

		clientset, err := pkg.GetClientset(kubeconfig, kubecontext)
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()
		for _, namespace := range namespaces {
			log.Infof("Processing namespace %s", pkg.NS(namespace))
			if skipDeployments {
				log.Infof("Skipping deployments in namespace %s", pkg.NS(namespace))
			} else {
				deployments, err := pkg.GetDeployments(ctx, clientset, namespace)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Found %s deployments in namespace %s", pkg.N(len(deployments.Items)), pkg.NS(namespace))
				downscaledDeployments, err := pkg.DownscaleDeployments(ctx, clientset, deployments, dryRun)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Downscaled %s deployments", pkg.N(downscaledDeployments))
			}

			if skipStatefulsets {
				log.Infof("Skipping statefulsets in namespace %s", pkg.NS(namespace))
			} else {
				statefulsets, err := pkg.GetStatefulSets(ctx, clientset, namespace)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Found %s statefulsets in namespace %s", pkg.N(len(statefulsets.Items)), pkg.NS(namespace))
				downscaledStatefulsets, err := pkg.DownscaleStatefulSets(ctx, clientset, statefulsets, dryRun)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Downscaled %s statefulsets", pkg.N(downscaledStatefulsets))
			}

			if skipDaemonsets {
				log.Infof("Skipping daemonsets in namespace %s", pkg.NS(namespace))
			} else {
				daemonsets, err := pkg.GetDaemonsets(ctx, clientset, namespace)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Found %s daemonsets in namespace %s", pkg.N(len(daemonsets.Items)), pkg.NS(namespace))
				downscaleDaemonsets, err := pkg.DownscaleDaemonsets(ctx, clientset, daemonsets, dryRun)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Downscaled %s daemonsets", pkg.N(downscaleDaemonsets))
			}
		}

		if wait && !dryRun {
			waitForResourcesOrFatal(ctx, clientset, true)
		}
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}

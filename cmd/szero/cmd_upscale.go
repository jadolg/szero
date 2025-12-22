package main

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/jadolg/szero/pkg"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:     "up",
	Short:   "Upscale all deployments/statefulsets/daemonsets in the desired namespaces to their original size",
	Example: "szero up -n default -n klum",
	Aliases: []string{"upscale"},
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
				upscaledDeployments, err := pkg.UpscaleDeployments(ctx, clientset, deployments, dryRun)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Upscaled %s deployments", pkg.N(upscaledDeployments))
			}

			if skipStatefulsets {
				log.Infof("Skipping statefulsets in namespace %s", pkg.NS(namespace))
			} else {
				statefulsets, err := pkg.GetStatefulSets(ctx, clientset, namespace)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Found %s statefulsets in namespace %s", pkg.N(len(statefulsets.Items)), pkg.NS(namespace))
				upscaledStatefulsets, err := pkg.UpscaleStatefulSets(ctx, clientset, statefulsets, dryRun)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Upscaled %s statefulsets", pkg.N(upscaledStatefulsets))
			}

			if skipDaemonsets {
				log.Infof("Skipping daemonsets in namespace %s", pkg.NS(namespace))
			} else {
				daemonsets, err := pkg.GetDaemonsets(ctx, clientset, namespace)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Found %s daemonsets in namespace %s", pkg.N(len(daemonsets.Items)), pkg.NS(namespace))
				upscaledDaemonsets, err := pkg.UpscaleDaemonsets(ctx, clientset, daemonsets, dryRun)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Upscaled %s daemonsets", pkg.N(upscaledDaemonsets))
			}
		}

		if wait && !dryRun {
			waitForResourcesOrFatal(ctx, clientset, false)
		}
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
}

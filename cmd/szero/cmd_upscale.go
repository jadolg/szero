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
				upscaledDeployments, err := pkg.UpscaleDeployments(ctx, clientset, deployments)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Upscaled %d deployments", upscaledDeployments)
			}

			if skipStatefulsets {
				log.Infof("Skipping statefulsets in namespace %s", namespace)
			} else {
				statefulsets, err := pkg.GetStatefulSets(ctx, clientset, namespace)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Found %d statefulsets in namespace %s", len(statefulsets.Items), namespace)
				upscaledStatefulsets, err := pkg.UpscaleStatefulSets(ctx, clientset, statefulsets)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Upscaled %d statefulsets", upscaledStatefulsets)
			}

			if skipDaemonsets {
				log.Infof("Skipping daemonsets in namespace %s", namespace)
			} else {
				daemonsets, err := pkg.GetDaemonsets(ctx, clientset, namespace)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Found %d daemonsets in namespace %s", len(daemonsets.Items), namespace)
				upscaledDaemonsets, err := pkg.UpscaleDaemonsets(ctx, clientset, daemonsets)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Upscaled %d daemonsets", upscaledDaemonsets)
			}
		}

		if wait {
			waitForResourcesOrFatal(ctx, clientset, false)
		}
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
}

package main

import (
	"context"
	"fmt"
	"os"

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
			fmt.Fprintln(os.Stderr, "⚠️  Running in dry-run mode, no changes will be made")
		}

		clientset, err := pkg.GetClientset(kubeconfig, kubecontext)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		printer := pkg.NewTreePrinter()
		ctx := context.Background()
		for _, namespace := range namespaces {
			result := pkg.NamespaceResult{
				Namespace: namespace,
			}

			// Deployments
			if skipDeployments {
				result.Deployments = pkg.ResourceGroup{Type: "Deployments", Skipped: true}
			} else {
				deployments, err := pkg.GetDeployments(ctx, clientset, namespace)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting deployments: %v\n", err)
					os.Exit(1)
				}
				deploymentInfos, err := pkg.DownscaleDeployments(ctx, clientset, deployments, dryRun)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error downscaling deployments: %v\n", err)
					os.Exit(1)
				}
				result.Deployments = pkg.ResourceGroup{
					Type:      "Deployments",
					Resources: deploymentInfos,
				}
			}

			// StatefulSets
			if skipStatefulsets {
				result.StatefulSets = pkg.ResourceGroup{Type: "StatefulSets", Skipped: true}
			} else {
				statefulsets, err := pkg.GetStatefulSets(ctx, clientset, namespace)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting statefulsets: %v\n", err)
					os.Exit(1)
				}
				statefulsetInfos, err := pkg.DownscaleStatefulSets(ctx, clientset, statefulsets, dryRun)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error downscaling statefulsets: %v\n", err)
					os.Exit(1)
				}
				result.StatefulSets = pkg.ResourceGroup{
					Type:      "StatefulSets",
					Resources: statefulsetInfos,
				}
			}

			// DaemonSets
			if skipDaemonsets {
				result.DaemonSets = pkg.ResourceGroup{Type: "DaemonSets", Skipped: true}
			} else {
				daemonsets, err := pkg.GetDaemonsets(ctx, clientset, namespace)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting daemonsets: %v\n", err)
					os.Exit(1)
				}
				daemonsetInfos, err := pkg.DownscaleDaemonsets(ctx, clientset, daemonsets, dryRun)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error downscaling daemonsets: %v\n", err)
					os.Exit(1)
				}
				result.DaemonSets = pkg.ResourceGroup{
					Type:      "DaemonSets",
					Resources: daemonsetInfos,
				}
			}

			printer.PrintNamespaceResult(result)
		}

		if wait && !dryRun {
			waitForResourcesOrFatal(ctx, clientset, true)
		}
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}

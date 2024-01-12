package main

import (
	"context"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

var (
	Version    = "dev"
	Commit     = "none"
	Date       = "unknown"
	BuiltBy    = "dirty hands"
	kubeconfig string
	namespaces []string
	rootCmd    = &cobra.Command{
		Use:   "szero",
		Short: "Completely downscale and upscale back deployments",
	}
)

func getDefaultKubeconfigPath() string {
	if os.Getenv("KUBECONFIG") != "" {
		return os.Getenv("KUBECONFIG")
	}
	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config")
	}
	return "kubeconfig"
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "k", getDefaultKubeconfigPath(), "Path to kubeconfig file")
	rootCmd.PersistentFlags().StringSliceVarP(&namespaces, "namespace", "n", []string{"default"}, "Kubernetes namespace")

	err := rootCmd.RegisterFlagCompletionFunc("namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		ctx := context.Background()
		clientset, err := getClientset(kubeconfig)
		if err != nil {
			log.Fatal(err)
		}
		return getNamespaces(ctx, clientset), cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		log.Fatal(err)
	}
}
func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

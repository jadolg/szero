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
	Version     = "dev"
	Commit      = "none"
	Date        = "unknown"
	BuiltBy     = "dirty hands"
	kubeconfig  string
	kubecontext string
	namespaces  []string
	rootCmd     = &cobra.Command{
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
	rootCmd.PersistentFlags().StringVarP(&kubecontext, "context", "c", getDefaultKubernetesContext(getDefaultKubeconfigPath()), "Kubernetes context")
	rootCmd.PersistentFlags().StringSliceVarP(&namespaces, "namespace", "n", []string{"default"}, "Kubernetes namespace")
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	err := rootCmd.RegisterFlagCompletionFunc("namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		ctx := context.Background()
		clientset, err := getClientset(kubeconfig, kubecontext)
		if err != nil {
			log.Fatal(err)
		}
		return getNamespaces(ctx, clientset), cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		log.Fatal(err)
	}
	err = rootCmd.RegisterFlagCompletionFunc("context", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		contexts, err := getContexts()
		if err != nil {
			log.Fatal(err)
		}
		return contexts, cobra.ShellCompDirectiveNoFileComp
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

package main

import (
	"context"
	"github.com/jadolg/szero/pkg"
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

	skipDaemonsets   bool
	skipStatefulsets bool
	skipDeployments  bool

	rootCmd = &cobra.Command{
		Use:   "szero",
		Short: "Temporarily scale down/up all deployments, statefulsets, and daemonsets in a namespace",
		Long:  "Downscale all deployments, statefulsets, and daemonsets in a namespace to 0 replicas and back to their previous state. Useful when you need to tear everything down and bring it back in a namespace.",
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
	defaultContext, defaultNamespace := pkg.GetDefaultKubernetesContextAndNamespace(getDefaultKubeconfigPath())
	rootCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "k", getDefaultKubeconfigPath(), "Path to kubeconfig file")
	rootCmd.PersistentFlags().StringVarP(&kubecontext, "context", "c", defaultContext, "Kubernetes context")
	rootCmd.PersistentFlags().StringSliceVarP(&namespaces, "namespace", "n", []string{defaultNamespace}, "Kubernetes namespace")

	rootCmd.PersistentFlags().BoolVarP(&skipDaemonsets, "skip-daemonsets", "d", false, "Skip daemonsets")
	rootCmd.PersistentFlags().BoolVarP(&skipStatefulsets, "skip-statefulsets", "s", false, "Skip statefulsets")
	rootCmd.PersistentFlags().BoolVarP(&skipDeployments, "skip-deployments", "p", false, "Skip deployments")

	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	err := rootCmd.RegisterFlagCompletionFunc("namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		ctx := context.Background()
		clientset, err := pkg.GetClientset(kubeconfig, kubecontext)
		if err != nil {
			log.Fatal(err)
		}
		return pkg.GetNamespaces(ctx, clientset), cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		log.Fatal(err)
	}
	err = rootCmd.RegisterFlagCompletionFunc("context", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		contexts, err := pkg.GetKubernetesContexts(kubeconfig)
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

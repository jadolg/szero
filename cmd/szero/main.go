package main

import (
	"context"
	"time"

	"github.com/jadolg/szero/pkg"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/samber/lo"
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

	wait    bool
	timeout time.Duration

	rootCmd = &cobra.Command{
		Use:   getApplicationName(),
		Short: "Temporarily scale down/up all deployments, statefulsets, and daemonsets in a namespace",
		Long:  "Downscale all deployments, statefulsets, and daemonsets in a namespace to 0 replicas and back to their previous state. Useful when you need to tear everything down and bring it back in a namespace.",
	}
)

func getApplicationName() string {
	if strings.HasPrefix(filepath.Base(os.Args[0]), "kubectl-") {
		return "kubectl-szero"
	}
	return "szero"
}

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

	rootCmd.PersistentFlags().BoolVarP(&wait, "wait", "w", false, "Wait for all resources to reconcile into the desired state")
	rootCmd.PersistentFlags().DurationVarP(&timeout, "timeout", "t", 5*time.Minute, "Timeout for waiting for resources to reconcile into the desired state")

	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	err := rootCmd.RegisterFlagCompletionFunc("namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		ctx := context.Background()
		clientset, err := pkg.GetClientset(kubeconfig, kubecontext)
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}
		_, unusedNamespaces := lo.Difference(namespaces, pkg.GetNamespaces(ctx, clientset))
		return unusedNamespaces, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		log.Fatal(err)
	}
	err = rootCmd.RegisterFlagCompletionFunc("context", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		contexts, err := pkg.GetKubernetesContexts(kubeconfig)
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}
		return contexts, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		log.Fatal(err)
	}
}

func setupLogs() {
	log.SetReportTimestamp(false)
	styles := log.DefaultStyles()
	styles.Levels[log.ErrorLevel] = lipgloss.NewStyle().
		SetString("😿").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("204")).
		Foreground(lipgloss.Color("0"))
	styles.Levels[log.InfoLevel] = lipgloss.NewStyle().
		SetString("🐱").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("86")).
		Foreground(lipgloss.Color("0"))
	styles.Levels[log.FatalLevel] = lipgloss.NewStyle().
		SetString("🙀").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("204")).
		Foreground(lipgloss.Color("0"))
	log.SetStyles(styles)
}

func main() {
	setupLogs()
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}

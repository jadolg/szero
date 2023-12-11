package main

import (
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

var (
	kubeconfig string
	namespace  string
	rootCmd    = &cobra.Command{
		Use:   "szero",
		Short: "Completely downscale and upscale back deployments",
	}
)

func getDefaultPath() string {
	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config")
	}
	return "kubeconfig"
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "k", getDefaultPath(), "Path to kubeconfig file")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
}
func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

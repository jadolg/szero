package pkg

import "k8s.io/client-go/tools/clientcmd"

func GetDefaultKubernetesContext(kubeconfig string) string {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{
			CurrentContext: "",
		}).RawConfig()
	if err != nil {
		return ""
	}
	return config.CurrentContext
}

func GetKubernetesContexts(kubeconfig string) ([]string, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{
			CurrentContext: "",
		}).RawConfig()
	if err != nil {
		return nil, err
	}
	contexts := make([]string, 0, len(config.Contexts))
	for key := range config.Contexts {
		contexts = append(contexts, key)
	}
	return contexts, nil
}

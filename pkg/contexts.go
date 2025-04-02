package pkg

import "k8s.io/client-go/tools/clientcmd"

func GetDefaultKubernetesContextAndNamespace(kubeconfig string) (string, string) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{
			CurrentContext: "",
		}).RawConfig()
	if err != nil {
		return "", "default"
	}
	namespace := "default"
	if _, found := config.Contexts[config.CurrentContext]; found && config.Contexts[config.CurrentContext].Namespace != "" {
		namespace = config.Contexts[config.CurrentContext].Namespace
	}
	return config.CurrentContext, namespace
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

package defaults

import "os"

const (
	PodNamespaceEnv     = "POD_NAMESPACE"
	DefaultPodNamespace = "gloo-mesh"
)

func GetPodNamespace() string {
	if podNamespace := os.Getenv(PodNamespaceEnv); podNamespace != "" {
		return podNamespace
	}
	return DefaultPodNamespace
}

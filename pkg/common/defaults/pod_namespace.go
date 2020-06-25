package defaults

import "os"

const (
	PodNamespaceEnv     = "POD_NAMESPACE"
	DefaultPodNamespace = "service-mesh-hub"
)

func GetPodNamespace() string {
	if podNamespace := os.Getenv(PodNamespaceEnv); podNamespace != "" {
		return podNamespace
	}
	return DefaultPodNamespace
}

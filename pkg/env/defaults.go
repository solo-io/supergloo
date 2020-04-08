package env

import "os"

const (
	EnvPodNamespace       = "POD_NAMESPACE"
	defaultWriteNamespace = "service-mesh-hub"
)

func GetWriteNamespace() string {
	writeNamespace := os.Getenv(EnvPodNamespace)
	if writeNamespace == "" {
		writeNamespace = defaultWriteNamespace
	}
	return writeNamespace
}

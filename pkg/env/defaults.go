package env

import "os"

const (
	EnvPodNamespace       = "POD_NAMESPACE"
	DefaultWriteNamespace = "service-mesh-hub"
)

func GetWriteNamespace() string {
	writeNamespace := os.Getenv(EnvPodNamespace)
	if writeNamespace == "" {
		writeNamespace = DefaultWriteNamespace
	}
	return writeNamespace
}

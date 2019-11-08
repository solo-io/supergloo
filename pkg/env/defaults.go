package env

import "os"

const (
	EnvPodNamespace       = "POD_NAMESPACE"
	DefaultWriteNamespace = "sm-marketplace"
)

func GetWriteNamespace() string {
	writeNamespace := os.Getenv(EnvPodNamespace)
	if writeNamespace == "" {
		writeNamespace = DefaultWriteNamespace
	}
	return writeNamespace
}
